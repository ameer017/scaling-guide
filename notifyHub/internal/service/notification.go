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

type NotificationService struct {
	repo     *repository.NotificationRepository
	producer *queue.Producer
	mailer   email.Mailer
}

func NewNotificationService(
	repo *repository.NotificationRepository,
	producer *queue.Producer,
	mailer email.Mailer,
) *NotificationService {
	return &NotificationService{repo: repo, producer: producer, mailer: mailer}
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

// ProcessDelivery is called by the worker after consuming a Kafka message.
func (s *NotificationService) ProcessDelivery(ctx context.Context, notificationID string) error {
	n, err := s.repo.GetByID(ctx, notificationID)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	// Idempotent: already delivered notifications are skipped.
	if n.Status == models.StatusSent {
		return nil
	}

	if s.mailer == nil {
		return fmt.Errorf("mailer is not configured")
	}

	if err := s.repo.UpdateStatus(ctx, n.ID, models.StatusProcessing, nil); err != nil {
		return err
	}

	if err := s.mailer.Send(ctx, n.Recipient, n.Subject, n.Body); err != nil {
		if updErr := s.repo.UpdateStatus(ctx, n.ID, models.StatusFailed, nil); updErr != nil {
			return fmt.Errorf("email send failed (%v) and status update failed: %w", err, updErr)
		}
		return fmt.Errorf("%w: %v", ErrDeliveryFailed, err)
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateStatus(ctx, n.ID, models.StatusSent, &now); err != nil {
		return err
	}
	return nil
}
