package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"notifyHub/internal/models"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func OpenPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	const q = `
		INSERT INTO notifications (id, recipient, subject, body, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, q,
		n.ID, n.Recipient, n.Subject, n.Body, n.Status, n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	const q = `
		SELECT id, recipient, subject, body, status, scheduled_at, sent_at, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, q, id)
	n, err := scanNotification(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return n, nil
}

func (r *NotificationRepository) List(ctx context.Context, limit int) ([]models.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	const q = `
		SELECT id, recipient, subject, body, status, scheduled_at, sent_at, created_at, updated_at
		FROM notifications
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	out := make([]models.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *n)
	}
	return out, rows.Err()
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id, status string, sentAt *time.Time) error {
	const q = `
		UPDATE notifications
		SET status = $2, sent_at = $3, updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, q, id, status, sentAt)
	if err != nil {
		return fmt.Errorf("update notification status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanNotification(row scannable) (*models.Notification, error) {
	var n models.Notification
	if err := row.Scan(
		&n.ID,
		&n.Recipient,
		&n.Subject,
		&n.Body,
		&n.Status,
		&n.ScheduledAt,
		&n.SentAt,
		&n.CreatedAt,
		&n.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &n, nil
}
