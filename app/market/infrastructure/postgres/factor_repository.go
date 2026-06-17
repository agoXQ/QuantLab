package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/lib/pq"
)

// FactorRepository persists Factor values.
type FactorRepository struct{ db *sql.DB }

// NewFactorRepository creates a new FactorRepository.
func NewFactorRepository(db *sql.DB) *FactorRepository { return &FactorRepository{db: db} }

// List implements factor.Repository.
func (r *FactorRepository) List(ctx context.Context, q factor.ListQuery) ([]*factor.Factor, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 1000
	}
	args := []any{strings.ToUpper(q.StockCode)}
	conditions := []string{"stock_code = $1"}
	if len(q.FactorNames) > 0 {
		args = append(args, pq.Array(q.FactorNames))
		conditions = append(conditions, fmt.Sprintf("factor_name = ANY($%d)", len(args)))
	}
	if !q.Range.Start.IsZero() {
		args = append(args, q.Range.Start)
		conditions = append(conditions, fmt.Sprintf("trade_date >= $%d", len(args)))
	}
	if !q.Range.End.IsZero() {
		args = append(args, q.Range.End)
		conditions = append(conditions, fmt.Sprintf("trade_date <= $%d", len(args)))
	}
	if q.DataVersion != "" {
		args = append(args, q.DataVersion)
		conditions = append(conditions, fmt.Sprintf("data_version = $%d", len(args)))
	}
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT stock_code, trade_date, factor_name,
		       COALESCE(factor_value, 0), COALESCE(data_version, '')
		FROM factor_data
		WHERE %s
		ORDER BY trade_date ASC, factor_name ASC
		LIMIT $%d`, strings.Join(conditions, " AND "), len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query factor_data: %w", err)
	}
	defer rows.Close()

	out := make([]*factor.Factor, 0, limit)
	for rows.Next() {
		var f factor.Factor
		if err := rows.Scan(&f.StockCode, &f.TradeDate, &f.FactorName, &f.FactorValue, &f.DataVersion); err != nil {
			return nil, fmt.Errorf("scan factor_data: %w", err)
		}
		out = append(out, &f)
	}
	return out, rows.Err()
}

// BulkUpsert implements factor.Repository.
func (r *FactorRepository) BulkUpsert(ctx context.Context, factors []*factor.Factor) error {
	if len(factors) == 0 {
		return nil
	}
	const stmt = `
		INSERT INTO factor_data (
			stock_code, trade_date, factor_name, factor_value, data_version
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (stock_code, trade_date, factor_name, data_version) DO UPDATE SET
			factor_value = EXCLUDED.factor_value`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, f := range factors {
		_, err := tx.ExecContext(ctx, stmt,
			strings.ToUpper(f.StockCode), f.TradeDate, f.FactorName, f.FactorValue, f.DataVersion,
		)
		if err != nil {
			return fmt.Errorf("upsert factor %s %s %s: %w", f.StockCode, f.TradeDate, f.FactorName, err)
		}
	}
	return tx.Commit()
}
