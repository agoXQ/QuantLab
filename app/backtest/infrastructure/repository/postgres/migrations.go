// Package postgres provides backtest repository implementations backed by
// PostgreSQL.
//
// Schema management is intentionally embedded so the service can self-
// bootstrap during local development. The canonical migration files live
// under app/backtest/migrations and stay the source of truth for any
// production deployment that runs a dedicated migration tool.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS backtest_job (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT,
    strategy_id     BIGINT,
    version_id      BIGINT,
    name            VARCHAR(255),
    formula         TEXT NOT NULL,
    universe        TEXT[] NOT NULL,
    benchmark       VARCHAR(64),
    data_version    VARCHAR(32),
    initial_capital NUMERIC(20,2) NOT NULL,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    status          VARCHAR(32) NOT NULL DEFAULT 'CREATED',
    error_message   TEXT,
    config          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_backtest_job_user
    ON backtest_job (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backtest_job_strategy
    ON backtest_job (strategy_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backtest_job_status
    ON backtest_job (status);

CREATE TABLE IF NOT EXISTS backtest_order (
    id              BIGSERIAL PRIMARY KEY,
    job_id          BIGINT NOT NULL REFERENCES backtest_job(id) ON DELETE CASCADE,
    stock_code      VARCHAR(32) NOT NULL,
    side            VARCHAR(8) NOT NULL,
    quantity        BIGINT NOT NULL,
    limit_price     NUMERIC(20,4),
    status          VARCHAR(16) NOT NULL,
    reason          VARCHAR(64),
    submitted_at    TIMESTAMPTZ NOT NULL,
    filled_at       TIMESTAMPTZ,
    filled_price    NUMERIC(20,4),
    filled_qty      BIGINT
);

CREATE INDEX IF NOT EXISTS idx_backtest_order_job
    ON backtest_order (job_id, submitted_at);

CREATE TABLE IF NOT EXISTS backtest_trade (
    id          BIGSERIAL PRIMARY KEY,
    job_id      BIGINT NOT NULL REFERENCES backtest_job(id) ON DELETE CASCADE,
    order_id    BIGINT NOT NULL REFERENCES backtest_order(id),
    stock_code  VARCHAR(32) NOT NULL,
    side        VARCHAR(8) NOT NULL,
    quantity    BIGINT NOT NULL,
    price       NUMERIC(20,4) NOT NULL,
    commission  NUMERIC(20,4) NOT NULL DEFAULT 0,
    stamp_duty  NUMERIC(20,4) NOT NULL DEFAULT 0,
    slippage    NUMERIC(20,4) NOT NULL DEFAULT 0,
    trade_time  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_backtest_trade_job
    ON backtest_trade (job_id, trade_time);

CREATE TABLE IF NOT EXISTS backtest_portfolio_snapshot (
    job_id      BIGINT NOT NULL REFERENCES backtest_job(id) ON DELETE CASCADE,
    trade_date  DATE NOT NULL,
    cash        NUMERIC(20,2) NOT NULL,
    market_value NUMERIC(20,2) NOT NULL,
    total_asset NUMERIC(20,2) NOT NULL,
    positions   JSONB NOT NULL DEFAULT '[]'::jsonb,
    PRIMARY KEY (job_id, trade_date)
);

CREATE TABLE IF NOT EXISTS backtest_report (
    job_id          BIGINT PRIMARY KEY REFERENCES backtest_job(id) ON DELETE CASCADE,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    initial_capital NUMERIC(20,2) NOT NULL,
    final_asset     NUMERIC(20,2) NOT NULL,
    total_return    NUMERIC(10,6) NOT NULL,
    annual_return   NUMERIC(10,6) NOT NULL,
    volatility      NUMERIC(10,6) NOT NULL,
    sharpe_ratio    NUMERIC(10,6) NOT NULL,
    max_drawdown    NUMERIC(10,6) NOT NULL,
    win_rate        NUMERIC(10,6) NOT NULL,
    trade_count     INT NOT NULL DEFAULT 0,
    equity_curve    JSONB NOT NULL DEFAULT '[]'::jsonb,
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// EnsureSchema creates the backtest schema if missing. It is safe to call
// repeatedly; every statement uses IF NOT EXISTS guards so the operation is
// idempotent across boots.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply backtest schema: %w", err)
	}
	return nil
}
