-- +migrate Up
CREATE TABLE IF NOT EXISTS delivery_logs (
    id              UUID PRIMARY KEY,
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL,
    response        TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL,
    attempt         INT  NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_delivery_logs_notification_id
    ON delivery_logs (notification_id, attempt);
