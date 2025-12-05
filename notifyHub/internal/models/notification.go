package models

import "time"

const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusSent       = "SENT"
	StatusFailed     = "FAILED"
)

// Notification is an email notification job.
type Notification struct {
	ID          string     `json:"id"`
	Recipient   string     `json:"recipient"`
	Subject     string     `json:"subject"`
	Body        string     `json:"body"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateNotificationRequest is the API payload to enqueue an email.
type CreateNotificationRequest struct {
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}
