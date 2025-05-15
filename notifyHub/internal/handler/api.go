package handler

import (
	"context"

	"notifyHub/internal/models"
)

// NotificationAPI is the service surface used by HTTP handlers.
type NotificationAPI interface {
	Create(ctx context.Context, req models.CreateNotificationRequest) (*models.Notification, error)
	Get(ctx context.Context, id string) (*models.Notification, error)
	List(ctx context.Context, limit int) ([]models.Notification, error)
	ListDeliveryLogs(ctx context.Context, notificationID string) ([]models.DeliveryLog, error)
	CreateTemplate(ctx context.Context, req models.CreateTemplateRequest) (*models.Template, error)
	ListTemplates(ctx context.Context, limit int) ([]models.Template, error)
}
