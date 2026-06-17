// Package postgres provides repository implementations backed by
// PostgreSQL/TimescaleDB.
//
// Schema management is intentionally embedded in this package so the service
// can self-bootstrap during local development. In production the platform is
// expected to apply the migration files under app/market/migrations through a
// dedicated migration tool.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS data_version (
    version     VARCHAR(32) PRIMARY KEY,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS security (
    id             BIGSERIAL PRIMARY KEY,
    stock_code     VARCHAR(32) NOT NULL,
    stock_name     VARCHAR(128),
    market         VARCHAR(8) NOT NULL DEFAULT 'CN',
    exchange       VARCHAR(32),
    asset_type     VARCHAR(16) NOT NULL DEFAULT 'STOCK',
    industry       VARCHAR(128),
    listing_date   DATE,
    delisting_date DATE,
    status         VARCHAR(16) NOT NULL DEFAULT 'LISTED',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_security_code ON security (stock_code);
CREATE INDEX IF NOT EXISTS idx_security_market ON security (market, asset_type);

CREATE TABLE IF NOT EXISTS trading_calendar (
    trade_date DATE PRIMARY KEY,
    is_open    BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS market_bar (
    stock_code   VARCHAR(32) NOT NULL,
    trade_date   DATE        NOT NULL,
    period       VARCHAR(8)  NOT NULL,
    open_price   DECIMAL(20,6),
    high_price   DECIMAL(20,6),
    low_price    DECIMAL(20,6),
    close_price  DECIMAL(20,6),
    volume       BIGINT,
    amount       DECIMAL(20,4),
    adj_factor   DECIMAL(20,6),
    data_version VARCHAR(32) NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_market_bar
    ON market_bar (stock_code, trade_date, period, data_version);
CREATE INDEX IF NOT EXISTS idx_market_bar_code_date
    ON market_bar (stock_code, trade_date DESC);

CREATE TABLE IF NOT EXISTS financial_statement (
    stock_code        VARCHAR(32) NOT NULL,
    report_date       DATE        NOT NULL,
    report_type       VARCHAR(16) NOT NULL,
    revenue           DECIMAL(20,4),
    net_profit        DECIMAL(20,4),
    total_assets      DECIMAL(20,4),
    total_liabilities DECIMAL(20,4),
    net_assets        DECIMAL(20,4),
    operating_cash_flow DECIMAL(20,4),
    investing_cash_flow DECIMAL(20,4),
    financing_cash_flow DECIMAL(20,4),
    data_version      VARCHAR(32) NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_financial_statement
    ON financial_statement (stock_code, report_date, report_type, data_version);

CREATE TABLE IF NOT EXISTS factor_data (
    stock_code   VARCHAR(32) NOT NULL,
    trade_date   DATE        NOT NULL,
    factor_name  VARCHAR(64) NOT NULL,
    factor_value DECIMAL(20,6),
    data_version VARCHAR(32) NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_factor_data
    ON factor_data (stock_code, trade_date, factor_name, data_version);

CREATE TABLE IF NOT EXISTS index_bar (
    index_code   VARCHAR(32) NOT NULL,
    trade_date   DATE        NOT NULL,
    open_price   DECIMAL(20,6),
    high_price   DECIMAL(20,6),
    low_price    DECIMAL(20,6),
    close_price  DECIMAL(20,6),
    volume       BIGINT,
    amount       DECIMAL(20,4),
    data_version VARCHAR(32) NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_index_bar
    ON index_bar (index_code, trade_date, data_version);
`

// hypertableStmts are best-effort statements to convert time-series tables to
// TimescaleDB hypertables. They are no-ops when the extension is not loaded.
var hypertableStmts = []string{
	`SELECT create_hypertable('market_bar',          'trade_date', if_not_exists => TRUE);`,
	`SELECT create_hypertable('financial_statement', 'report_date', if_not_exists => TRUE);`,
	`SELECT create_hypertable('factor_data',         'trade_date', if_not_exists => TRUE);`,
	`SELECT create_hypertable('index_bar',           'trade_date', if_not_exists => TRUE);`,
}

// EnsureSchema creates the market schema if it is missing and, when the
// TimescaleDB extension is available, converts time-series tables into
// hypertables. The function is safe to call repeatedly.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply market schema: %w", err)
	}
	for _, stmt := range hypertableStmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			if isHypertableSkippable(err) {
				continue
			}
			return fmt.Errorf("create hypertable: %w", err)
		}
	}
	return nil
}

func isHypertableSkippable(err error) bool {
	if err == nil {
		return true
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "function create_hypertable"):
		return true
	case strings.Contains(msg, "extension \"timescaledb\""):
		return true
	case strings.Contains(msg, "already a hypertable"):
		return true
	}
	return errors.Is(err, sql.ErrConnDone)
}
