// Package postgres provides Strategy repository implementations backed
// by PostgreSQL.
//
// The schema is embedded so the service can self-bootstrap during
// local development. Production deployments still apply migrations via
// the platform's migration tool; the embedded SQL stays compatible
// because every statement uses IF NOT EXISTS / ADD COLUMN IF NOT
// EXISTS guards.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS strategy (
    id                  BIGSERIAL PRIMARY KEY,
    author_id           BIGINT,
    title               VARCHAR(256) NOT NULL,
    description         TEXT,
    category            VARCHAR(64),
    tags                TEXT[] NOT NULL DEFAULT '{}',
    current_version_id  BIGINT,
    status              VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    visibility          VARCHAR(32) NOT NULL DEFAULT 'PRIVATE',
    source_strategy_id  BIGINT,
    view_count          BIGINT NOT NULL DEFAULT 0,
    favorite_count      BIGINT NOT NULL DEFAULT 0,
    fork_count          BIGINT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at        TIMESTAMPTZ,
    archived_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_strategy_author
    ON strategy (author_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_strategy_status
    ON strategy (status);
CREATE INDEX IF NOT EXISTS idx_strategy_visibility
    ON strategy (visibility, created_at DESC);

CREATE TABLE IF NOT EXISTS strategy_version (
    id              BIGSERIAL PRIMARY KEY,
    strategy_id     BIGINT NOT NULL REFERENCES strategy(id) ON DELETE CASCADE,
    version_no      VARCHAR(32) NOT NULL,
    formula_text    TEXT NOT NULL,
    buy_rule        TEXT,
    sell_rule       TEXT,
    risk_rule       TEXT,
    position_rule   TEXT,
    rebalance_rule  TEXT,
    change_log      TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_version_strategy
    ON strategy_version (strategy_id, id DESC);

CREATE TABLE IF NOT EXISTS strategy_fork (
    id                  BIGSERIAL PRIMARY KEY,
    source_strategy_id  BIGINT NOT NULL,
    target_strategy_id  BIGINT NOT NULL,
    creator_id          BIGINT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_fork_source
    ON strategy_fork (source_strategy_id, created_at DESC);
`

// EnsureSchema creates the Strategy schema if missing. It is safe to
// call repeatedly; every statement uses IF NOT EXISTS guards so the
// operation is idempotent across boots.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply strategy schema: %w", err)
	}
	return nil
}
