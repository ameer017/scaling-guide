package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"notifyHub/internal/models"
)

type DeliveryLogRepository struct {
	pool *pgxpool.Pool
}

func NewDeliveryLogRepository(pool *pgxpool.Pool) *DeliveryLogRepository {
	return &DeliveryLogRepository{pool: pool}
}

func (r *DeliveryLogRepository) Create(ctx context.Context, log *models.DeliveryLog) error {
	const q = `
		INSERT INTO delivery_logs (id, notification_id, provider, response, status, attempt, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, q,
		log.ID, log.NotificationID, log.Provider, log.Response, log.Status, log.Attempt, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert delivery log: %w", err)
	}
	return nil
}

func (r *DeliveryLogRepository) CountByNotification(ctx context.Context, notificationID string) (int, error) {
	const q = `SELECT COUNT(*) FROM delivery_logs WHERE notification_id = $1`
	var count int
	if err := r.pool.QueryRow(ctx, q, notificationID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count delivery logs: %w", err)
	}
	return count, nil
}

func (r *DeliveryLogRepository) ListByNotification(ctx context.Context, notificationID string) ([]models.DeliveryLog, error) {
	const q = `
		SELECT id, notification_id, provider, response, status, attempt, created_at
		FROM delivery_logs
		WHERE notification_id = $1
		ORDER BY attempt ASC
	`
	rows, err := r.pool.Query(ctx, q, notificationID)
	if err != nil {
		return nil, fmt.Errorf("list delivery logs: %w", err)
	}
	defer rows.Close()

	out := make([]models.DeliveryLog, 0)
	for rows.Next() {
		var log models.DeliveryLog
		if err := rows.Scan(
			&log.ID,
			&log.NotificationID,
			&log.Provider,
			&log.Response,
			&log.Status,
			&log.Attempt,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, log)
	}
	return out, rows.Err()
}
