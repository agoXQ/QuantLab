// Package postgres provides Notification repository implementations
// backed by PostgreSQL. The schema is embedded so the service can
// self-bootstrap during local development; production deployments
// still apply migrations via the platform tool, but every statement
// uses IF NOT EXISTS guards so the bootstrap stays idempotent.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS notification (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL,
    type        INT         NOT NULL,
    title       VARCHAR(256) NOT NULL,
    content     TEXT        NOT NULL DEFAULT '',
    status      INT         NOT NULL DEFAULT 1,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_user_created
    ON notification (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notification_user_status
    ON notification (user_id, status);

CREATE TABLE IF NOT EXISTS notification_preference (
    user_id          BIGINT PRIMARY KEY,
    in_app_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    email_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
    webhook_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    push_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_subscription (
    id             BIGSERIAL PRIMARY KEY,
    subscriber_id  BIGINT      NOT NULL,
    object_type    VARCHAR(64) NOT NULL,
    object_id      BIGINT      NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_notification_subscription_triple
    ON notification_subscription (subscriber_id, object_type, object_id);
CREATE INDEX IF NOT EXISTS idx_notification_subscription_subscriber
    ON notification_subscription (subscriber_id, created_at DESC);
`

// EnsureSchema creates the Notification schema if missing.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply notification schema: %w", err)
	}
	return nil
}
