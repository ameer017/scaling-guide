-- +migrate Up
CREATE TABLE IF NOT EXISTS templates (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    subject    TEXT NOT NULL,
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_scheduled_due
    ON notifications (scheduled_at)
    WHERE status = 'SCHEDULED';
