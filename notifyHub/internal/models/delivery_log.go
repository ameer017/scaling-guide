package models

import "time"

const (
	DeliveryStatusSuccess = "SUCCESS"
	DeliveryStatusFailure = "FAILURE"
)

// DeliveryLog records one send attempt for a notification.
type DeliveryLog struct {
	ID             string    `json:"id"`
	NotificationID string    `json:"notification_id"`
	Provider       string    `json:"provider"`
	Response       string    `json:"response"`
	Status         string    `json:"status"`
	Attempt        int       `json:"attempt"`
	CreatedAt      time.Time `json:"created_at"`
}
