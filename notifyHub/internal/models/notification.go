package models

import "time"

const (
	StatusPending    = "PENDING"
	StatusScheduled  = "SCHEDULED"
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
// Provide subject+body, or template_id (+ optional variables).
type CreateNotificationRequest struct {
	Recipient   string            `json:"recipient"`
	Subject     string            `json:"subject,omitempty"`
	Body        string            `json:"body,omitempty"`
	TemplateID  string            `json:"template_id,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
}

// Template is a reusable email subject/body with {{placeholders}}.
type Template struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTemplateRequest is the API payload to create a template.
type CreateTemplateRequest struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
