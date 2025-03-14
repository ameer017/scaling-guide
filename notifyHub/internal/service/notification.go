package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"notifyHub/internal/email"
	"notifyHub/internal/models"
	"notifyHub/internal/queue"
	"notifyHub/internal/repository"
)

const (
	OutcomeSent            = "sent"
	OutcomeRetryScheduled  = "retry_scheduled"
	OutcomeFailedPermanent = "failed_permanent"
	OutcomeSkipped         = "skipped"
	defaultEmailProvider   = "gmail-smtp"
	maxBackoff             = 30 * time.Second
)

type Deps struct {
	Repo         *repository.NotificationRepository
	Templates    *repository.TemplateRepository
	Logs         *repository.DeliveryLogRepository
	Producer     *queue.Producer
	DLQ          *queue.Producer
	Mailer       email.Mailer
	MaxAttempts  int
	BackoffBase  time.Duration
	Provider     string
}

type NotificationService struct {
	repo        *repository.NotificationRepository
	templates   *repository.TemplateRepository
	logs        *repository.DeliveryLogRepository
	producer    *queue.Producer
	dlq         *queue.Producer
	mailer      email.Mailer
	maxAttempts int
	backoffBase time.Duration
	provider    string
}

func NewNotificationService(d Deps) *NotificationService {
	maxAttempts := d.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	backoff := d.BackoffBase
	if backoff <= 0 {
		backoff = 2 * time.Second
	}
	provider := d.Provider
	if provider == "" {
		provider = defaultEmailProvider
	}
	return &NotificationService{
		repo:        d.Repo,
		templates:   d.Templates,
		logs:        d.Logs,
		producer:    d.Producer,
		dlq:         d.DLQ,
		mailer:      d.Mailer,
		maxAttempts: maxAttempts,
		backoffBase: backoff,
		provider:    provider,
	}
}

func (s *NotificationService) Create(ctx context.Context, req models.CreateNotificationRequest) (*models.Notification, error) {
	req.Recipient = strings.TrimSpace(req.Recipient)
	if req.Recipient == "" {
		return nil, fmt.Errorf("%w: recipient is required", ErrInvalidInput)
	}

	subject, body, err := s.resolveContent(ctx, req)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	status := models.StatusPending
	var scheduledAt *time.Time

	if req.ScheduledAt != nil {
		at := req.ScheduledAt.UTC()
		if at.After(now) {
			status = models.StatusScheduled
			scheduledAt = &at
		}
	}

	n := &models.Notification{
		ID:          uuid.NewString(),
		Recipient:   req.Recipient,
		Subject:     subject,
		Body:        body,
		Status:      status,
		ScheduledAt: scheduledAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}

	// Future sends wait for the scheduler; immediate sends go to Kafka now.
	if status == models.StatusPending {
		if err := s.producer.Publish(ctx, n.ID); err != nil {
			return nil, fmt.Errorf("notification saved but kafka publish failed: %w", err)
		}
	}

	return n, nil
}

func (s *NotificationService) resolveContent(ctx context.Context, req models.CreateNotificationRequest) (subject, body string, err error) {
	templateID := strings.TrimSpace(req.TemplateID)
	if templateID != "" {
		if s.templates == nil {
			return "", "", fmt.Errorf("template repository is not configured")
		}
		t, getErr := s.templates.GetByID(ctx, templateID)
		if getErr != nil {
			if getErr == repository.ErrNotFound {
				return "", "", fmt.Errorf("%w: template not found", ErrNotFound)
			}
			return "", "", getErr
		}
		subject = RenderTemplate(t.Subject, req.Variables)
		body = RenderTemplate(t.Body, req.Variables)
		return subject, body, nil
	}

	subject = strings.TrimSpace(req.Subject)
	body = strings.TrimSpace(req.Body)
	if subject == "" {
		return "", "", fmt.Errorf("%w: subject is required (or provide template_id)", ErrInvalidInput)
	}
	if body == "" {
		return "", "", fmt.Errorf("%w: body is required (or provide template_id)", ErrInvalidInput)
	}
	return subject, body, nil
}

func (s *NotificationService) CreateTemplate(ctx context.Context, req models.CreateTemplateRequest) (*models.Template, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Subject = strings.TrimSpace(req.Subject)
	req.Body = strings.TrimSpace(req.Body)
	if req.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if req.Subject == "" {
		return nil, fmt.Errorf("%w: subject is required", ErrInvalidInput)
	}
	if req.Body == "" {
		return nil, fmt.Errorf("%w: body is required", ErrInvalidInput)
	}
	if s.templates == nil {
		return nil, fmt.Errorf("template repository is not configured")
	}

	t := &models.Template{
		ID:        uuid.NewString(),
		Name:      req.Name,
		Subject:   req.Subject,
		Body:      req.Body,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.templates.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *NotificationService) ListTemplates(ctx context.Context, limit int) ([]models.Template, error) {
	if s.templates == nil {
		return nil, fmt.Errorf("template repository is not configured")
	}
	return s.templates.List(ctx, limit)
}

func (s *NotificationService) Get(ctx context.Context, id string) (*models.Notification, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return n, nil
}

func (s *NotificationService) List(ctx context.Context, limit int) ([]models.Notification, error) {
	return s.repo.List(ctx, limit)
}

func (s *NotificationService) ListDeliveryLogs(ctx context.Context, notificationID string) ([]models.DeliveryLog, error) {
	if _, err := s.Get(ctx, notificationID); err != nil {
		return nil, err
	}
	return s.logs.ListByNotification(ctx, notificationID)
}

// ProcessDelivery handles one Kafka delivery attempt.
func (s *NotificationService) ProcessDelivery(ctx context.Context, notificationID string) (string, error) {
	n, err := s.repo.GetByID(ctx, notificationID)
	if err != nil {
		if err == repository.ErrNotFound {
			_ = s.publishDLQ(ctx, notificationID, "notification not found", 0)
			return OutcomeFailedPermanent, ErrNotFound
		}
		return "", err
	}

	if n.Status == models.StatusSent || n.Status == models.StatusFailed {
		return OutcomeSkipped, nil
	}
	if n.Status == models.StatusScheduled {
		// Not due yet — scheduler owns these. Skip if a message arrived early.
		return OutcomeSkipped, nil
	}

	if s.mailer == nil {
		return "", fmt.Errorf("mailer is not configured")
	}
	if s.logs == nil {
		return "", fmt.Errorf("delivery log repository is not configured")
	}

	priorAttempts, err := s.logs.CountByNotification(ctx, n.ID)
	if err != nil {
		return "", err
	}
	attempt := priorAttempts + 1

	if priorAttempts >= s.maxAttempts {
		if err := s.repo.UpdateStatus(ctx, n.ID, models.StatusFailed, nil); err != nil {
			return "", err
		}
		if err := s.publishDLQ(ctx, n.ID, "max attempts already exhausted", priorAttempts); err != nil {
			return "", err
		}
		return OutcomeFailedPermanent, nil
	}

	if err := s.repo.UpdateStatus(ctx, n.ID, models.StatusProcessing, nil); err != nil {
		return "", err
	}

	sendErr := s.mailer.Send(ctx, n.Recipient, n.Subject, n.Body)
	now := time.Now().UTC()

	logEntry := &models.DeliveryLog{
		ID:             uuid.NewString(),
		NotificationID: n.ID,
		Provider:       s.provider,
		Attempt:        attempt,
		CreatedAt:      now,
	}

	if sendErr == nil {
		logEntry.Status = models.DeliveryStatusSuccess
		logEntry.Response = "ok"
		if err := s.logs.Create(ctx, logEntry); err != nil {
			return "", err
		}
		if err := s.repo.UpdateStatus(ctx, n.ID, models.StatusSent, &now); err != nil {
			return "", err
		}
		return OutcomeSent, nil
	}

	logEntry.Status = models.DeliveryStatusFailure
	logEntry.Response = sendErr.Error()
	if err := s.logs.Create(ctx, logEntry); err != nil {
		return "", fmt.Errorf("email send failed (%v) and delivery log failed: %w", sendErr, err)
	}

	if attempt < s.maxAttempts {
		if err := s.repo.UpdateStatus(ctx, n.ID, models.StatusPending, nil); err != nil {
			return "", err
		}
		if err := s.waitBackoff(ctx, attempt); err != nil {
			return "", err
		}
		if err := s.producer.Publish(ctx, n.ID); err != nil {
			return "", fmt.Errorf("retry publish failed after attempt %d: %w", attempt, err)
		}
		return OutcomeRetryScheduled, fmt.Errorf("%w: attempt %d/%d: %v", ErrRetryScheduled, attempt, s.maxAttempts, sendErr)
	}

	if err := s.repo.UpdateStatus(ctx, n.ID, models.StatusFailed, nil); err != nil {
		return "", err
	}
	if err := s.publishDLQ(ctx, n.ID, sendErr.Error(), attempt); err != nil {
		return "", err
	}
	return OutcomeFailedPermanent, fmt.Errorf("%w: attempt %d/%d: %v", ErrDeliveryFailed, attempt, s.maxAttempts, sendErr)
}

func (s *NotificationService) waitBackoff(ctx context.Context, attempt int) error {
	shift := attempt - 1
	if shift < 0 {
		shift = 0
	}
	backoff := s.backoffBase * time.Duration(1<<shift)
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *NotificationService) publishDLQ(ctx context.Context, notificationID, reason string, attempt int) error {
	if s.dlq == nil {
		return fmt.Errorf("dlq producer is not configured")
	}
	return s.dlq.PublishDLQ(ctx, queue.DLQMessage{
		NotificationID: notificationID,
		Reason:         reason,
		Attempt:        attempt,
	})
}
