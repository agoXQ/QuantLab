package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/agoXQ/QuantLab/app/backtest/domain/order"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// OrderRepository persists Order entities in PostgreSQL.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository wires the repository.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// BulkInsert writes all orders inside a single transaction so the run
// produces a coherent slice even when the engine commits piecemeal.
func (r *OrderRepository) BulkInsert(ctx context.Context, orders []*order.Order) error {
	if len(orders) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const stmt = `
		INSERT INTO backtest_order (
			job_id, stock_code, side, quantity, limit_price, status, reason,
			submitted_at, filled_at, filled_price, filled_qty
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	for _, o := range orders {
		if o == nil {
			continue
		}
		var id int64
		err := tx.QueryRowContext(ctx, stmt,
			o.JobID, o.StockCode, string(o.Side), o.Quantity,
			nullableFloat(o.LimitPrice), string(o.Status), nullableString(o.Reason),
			o.SubmittedAt, o.FilledAt, nullableFloat(o.FilledPrice), nullableInt64(o.FilledQty),
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("insert backtest_order: %w", err)
		}
		o.ID = id
	}
	return tx.Commit()
}

// ListByJob returns all orders attached to a job ordered by submission time.
func (r *OrderRepository) ListByJob(ctx context.Context, jobID int64) ([]*order.Order, error) {
	const q = `
		SELECT id, job_id, stock_code, side, quantity,
		       COALESCE(limit_price, 0), status, COALESCE(reason, ''),
		       submitted_at, filled_at,
		       COALESCE(filled_price, 0), COALESCE(filled_qty, 0)
		FROM backtest_order
		WHERE job_id = $1
		ORDER BY submitted_at, id`
	rows, err := r.db.QueryContext(ctx, q, jobID)
	if err != nil {
		return nil, fmt.Errorf("query backtest_order: %w", err)
	}
	defer rows.Close()
	out := make([]*order.Order, 0)
	for rows.Next() {
		var (
			o        order.Order
			side     string
			status   string
			filledAt sql.NullTime
		)
		if err := rows.Scan(
			&o.ID, &o.JobID, &o.StockCode, &side, &o.Quantity,
			&o.LimitPrice, &status, &o.Reason,
			&o.SubmittedAt, &filledAt,
			&o.FilledPrice, &o.FilledQty,
		); err != nil {
			return nil, fmt.Errorf("scan backtest_order: %w", err)
		}
		o.Side = valueobject.OrderSide(side)
		o.Status = valueobject.OrderStatus(status)
		if filledAt.Valid {
			t := filledAt.Time
			o.FilledAt = &t
		}
		out = append(out, &o)
	}
	return out, rows.Err()
}

func nullableFloat(v float64) any {
	if v == 0 {
		return nil
	}
	return v
}
