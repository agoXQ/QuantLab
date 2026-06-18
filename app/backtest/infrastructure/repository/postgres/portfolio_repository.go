package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
)

// PortfolioRepository persists Portfolio Snapshots in PostgreSQL.
type PortfolioRepository struct {
	db *sql.DB
}

// NewPortfolioRepository wires the repository.
func NewPortfolioRepository(db *sql.DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

// BulkInsertSnapshots upserts on (job_id, trade_date) so reruns of the same
// job replace existing snapshots cleanly.
func (r *PortfolioRepository) BulkInsertSnapshots(ctx context.Context, snapshots []portfolio.Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const stmt = `
		INSERT INTO backtest_portfolio_snapshot (
			job_id, trade_date, cash, market_value, total_asset, positions
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (job_id, trade_date) DO UPDATE SET
			cash         = EXCLUDED.cash,
			market_value = EXCLUDED.market_value,
			total_asset  = EXCLUDED.total_asset,
			positions    = EXCLUDED.positions`
	for _, s := range snapshots {
		positions := s.Positions
		if positions == nil {
			positions = []portfolio.Position{}
		}
		payload, err := json.Marshal(positions)
		if err != nil {
			return fmt.Errorf("marshal positions: %w", err)
		}
		if _, err := tx.ExecContext(ctx, stmt,
			s.JobID, s.TradeDate, s.Cash, s.MarketValue, s.TotalAsset, payload,
		); err != nil {
			return fmt.Errorf("upsert backtest_portfolio_snapshot: %w", err)
		}
	}
	return tx.Commit()
}

// ListSnapshots returns snapshots ordered by trade date.
func (r *PortfolioRepository) ListSnapshots(ctx context.Context, jobID int64) ([]portfolio.Snapshot, error) {
	const q = `
		SELECT job_id, trade_date, cash, market_value, total_asset, positions
		FROM backtest_portfolio_snapshot
		WHERE job_id = $1
		ORDER BY trade_date`
	rows, err := r.db.QueryContext(ctx, q, jobID)
	if err != nil {
		return nil, fmt.Errorf("query backtest_portfolio_snapshot: %w", err)
	}
	defer rows.Close()
	out := make([]portfolio.Snapshot, 0)
	for rows.Next() {
		var (
			s       portfolio.Snapshot
			payload []byte
		)
		if err := rows.Scan(
			&s.JobID, &s.TradeDate, &s.Cash, &s.MarketValue, &s.TotalAsset, &payload,
		); err != nil {
			return nil, fmt.Errorf("scan backtest_portfolio_snapshot: %w", err)
		}
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &s.Positions); err != nil {
				return nil, fmt.Errorf("unmarshal positions: %w", err)
			}
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
