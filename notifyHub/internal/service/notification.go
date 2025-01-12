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
	OutcomeSent             = "sent"
	OutcomeRetryScheduled   = "retry_scheduled"
	OutcomeFailedPermanent  = "failed_permanent"
	OutcomeSkipped          = "skipped"
	defaultEmailProvider    = "gmail-smtp"
	maxBackoff              = 30 * time.Second
)

type Deps struct {
	Repo        *repository.NotificationRepository
	Logs        *repository.DeliveryLogRepository
	Producer    *queue.Producer
	DLQ         *queue.Producer
	Mailer      email.Mailer
	MaxAttempts int
	BackoffBase time.Duration
	Provider    string
}

type NotificationService struct {
	repo        *repository.NotificationRepository
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
	req.Subject = strings.TrimSpace(req.Subject)
	if req.Recipient == "" {
		return nil, fmt.Errorf("%w: recipient is required", ErrInvalidInput)
	}
	if req.Subject == "" {
		return nil, fmt.Errorf("%w: subject is required", ErrInvalidInput)
	}
	if strings.TrimSpace(req.Body) == "" {
		return nil, fmt.Errorf("%w: body is required", ErrInvalidInput)
	}

	now := time.Now().UTC()
	n := &models.Notification{
		ID:        uuid.NewString(),
		Recipient: req.Recipient,
		Subject:   req.Subject,
		Body:      req.Body,
		Status:    models.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}

	// Persist first, then publish. If publish fails, the row stays PENDING for later recovery.
	if err := s.producer.Publish(ctx, n.ID); err != nil {
		return nil, fmt.Errorf("notification saved but kafka publish failed: %w", err)
	}

	return n, nil
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
// On retryable failure it waits (exponential backoff), re-publishes to the send topic, and returns
// OutcomeRetryScheduled so the worker can commit the current offset.
// On permanent failure it marks FAILED, publishes to the DLQ, and returns OutcomeFailedPermanent.
func (s *NotificationService) ProcessDelivery(ctx context.Context, notificationID string) (string, error) {
	n, err := s.repo.GetByID(ctx, notificationID)
	if err != nil {
		if err == repository.ErrNotFound {
			_ = s.publishDLQ(ctx, notificationID, "notification not found", 0)
			return OutcomeFailedPermanent, ErrNotFound
		}
		return "", err
	}

	if n.Status == models.StatusSent {
		return OutcomeSkipped, nil
	}
	if n.Status == models.StatusFailed {
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
