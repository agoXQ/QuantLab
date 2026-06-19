// Package postgres provides User repository implementations backed by
// PostgreSQL. The schema is embedded so the service can self-bootstrap
// during local development; production deployments still apply
// migrations via the platform's migration tool, but the embedded SQL
// stays compatible because every statement uses IF NOT EXISTS guards.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS app_user (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(64)  NOT NULL,
    email           VARCHAR(256) NOT NULL,
    password_hash   VARCHAR(256) NOT NULL,
    avatar          TEXT,
    bio             TEXT,
    nickname        VARCHAR(128),
    location        VARCHAR(128),
    status          INT NOT NULL DEFAULT 1,
    creator_status  INT NOT NULL DEFAULT 0,
    verified_status INT NOT NULL DEFAULT 0,
    membership_tier VARCHAR(32) NOT NULL DEFAULT 'FREE',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_app_user_email
    ON app_user (LOWER(email));
CREATE UNIQUE INDEX IF NOT EXISTS uniq_app_user_username
    ON app_user (LOWER(username));

CREATE TABLE IF NOT EXISTS user_follow (
    id           BIGSERIAL PRIMARY KEY,
    follower_id  BIGINT NOT NULL,
    followee_id  BIGINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_user_follow_pair
    ON user_follow (follower_id, followee_id);
CREATE INDEX IF NOT EXISTS idx_user_follow_followee
    ON user_follow (followee_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_follow_follower
    ON user_follow (follower_id, created_at DESC);
`

// EnsureSchema creates the User schema if missing. Safe to call
// repeatedly; every statement is idempotent across boots.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply user schema: %w", err)
	}
	return nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
