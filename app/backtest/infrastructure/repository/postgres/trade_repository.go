package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// TradeRepository persists Trade entities in PostgreSQL.
type TradeRepository struct {
	db *sql.DB
}

// NewTradeRepository wires the repository.
func NewTradeRepository(db *sql.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

// BulkInsert writes the trades emitted by a single run inside one tx.
func (r *TradeRepository) BulkInsert(ctx context.Context, trades []*trade.Trade) error {
	if len(trades) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const stmt = `
		INSERT INTO backtest_trade (
			job_id, order_id, stock_code, side, quantity, price,
			commission, stamp_duty, slippage, trade_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`
	for _, t := range trades {
		if t == nil {
			continue
		}
		var id int64
		err := tx.QueryRowContext(ctx, stmt,
			t.JobID, t.OrderID, t.StockCode, string(t.Side), t.Quantity, t.Price,
			t.Commission, t.StampDuty, t.Slippage, t.TradeTime,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("insert backtest_trade: %w", err)
		}
		t.ID = id
	}
	return tx.Commit()
}

// ListByJob returns trades for a job ordered by trade time.
func (r *TradeRepository) ListByJob(ctx context.Context, jobID int64) ([]*trade.Trade, error) {
	const q = `
		SELECT id, job_id, order_id, stock_code, side, quantity, price,
		       commission, stamp_duty, slippage, trade_time
		FROM backtest_trade
		WHERE job_id = $1
		ORDER BY trade_time, id`
	rows, err := r.db.QueryContext(ctx, q, jobID)
	if err != nil {
		return nil, fmt.Errorf("query backtest_trade: %w", err)
	}
	defer rows.Close()
	out := make([]*trade.Trade, 0)
	for rows.Next() {
		var (
			t    trade.Trade
			side string
		)
		if err := rows.Scan(
			&t.ID, &t.JobID, &t.OrderID, &t.StockCode, &side, &t.Quantity, &t.Price,
			&t.Commission, &t.StampDuty, &t.Slippage, &t.TradeTime,
		); err != nil {
			return nil, fmt.Errorf("scan backtest_trade: %w", err)
		}
		t.Side = valueobject.OrderSide(side)
		out = append(out, &t)
	}
	return out, rows.Err()
}
