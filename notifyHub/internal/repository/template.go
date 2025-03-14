package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"notifyHub/internal/models"
)

type TemplateRepository struct {
	pool *pgxpool.Pool
}

func NewTemplateRepository(pool *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{pool: pool}
}

func (r *TemplateRepository) Create(ctx context.Context, t *models.Template) error {
	const q = `
		INSERT INTO templates (id, name, subject, body, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, q, t.ID, t.Name, t.Subject, t.Body, t.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert template: %w", err)
	}
	return nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*models.Template, error) {
	const q = `
		SELECT id, name, subject, body, created_at
		FROM templates
		WHERE id = $1
	`
	var t models.Template
	err := r.pool.QueryRow(ctx, q, id).Scan(&t.ID, &t.Name, &t.Subject, &t.Body, &t.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepository) List(ctx context.Context, limit int) ([]models.Template, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	const q = `
		SELECT id, name, subject, body, created_at
		FROM templates
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	out := make([]models.Template, 0)
	for rows.Next() {
		var t models.Template
		if err := rows.Scan(&t.ID, &t.Name, &t.Subject, &t.Body, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
