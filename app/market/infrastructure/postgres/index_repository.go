package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
)

// IndexBarRepository persists IndexBar entities.
type IndexBarRepository struct{ db *sql.DB }

// NewIndexBarRepository creates a new IndexBarRepository.
func NewIndexBarRepository(db *sql.DB) *IndexBarRepository { return &IndexBarRepository{db: db} }

// List implements indexbar.Repository.
func (r *IndexBarRepository) List(ctx context.Context, q indexbar.RangeQuery) ([]*indexbar.IndexBar, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 1000
	}
	args := []any{strings.ToUpper(q.IndexCode)}
	conditions := []string{"index_code = $1"}
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
		SELECT index_code, trade_date,
		       COALESCE(open_price, 0), COALESCE(high_price, 0),
		       COALESCE(low_price, 0), COALESCE(close_price, 0),
		       COALESCE(volume, 0), COALESCE(amount, 0),
		       COALESCE(data_version, '')
		FROM index_bar
		WHERE %s
		ORDER BY trade_date ASC
		LIMIT $%d`, strings.Join(conditions, " AND "), len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query index_bar: %w", err)
	}
	defer rows.Close()

	out := make([]*indexbar.IndexBar, 0, limit)
	for rows.Next() {
		var b indexbar.IndexBar
		if err := rows.Scan(
			&b.IndexCode, &b.TradeDate, &b.Open, &b.High, &b.Low, &b.Close,
			&b.Volume, &b.Amount, &b.DataVersion,
		); err != nil {
			return nil, fmt.Errorf("scan index_bar: %w", err)
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}

// BulkUpsert implements indexbar.Repository.
func (r *IndexBarRepository) BulkUpsert(ctx context.Context, bars []*indexbar.IndexBar) error {
	if len(bars) == 0 {
		return nil
	}
	const stmt = `
		INSERT INTO index_bar (
			index_code, trade_date, open_price, high_price, low_price, close_price,
			volume, amount, data_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (index_code, trade_date, data_version) DO UPDATE SET
			open_price  = EXCLUDED.open_price,
			high_price  = EXCLUDED.high_price,
			low_price   = EXCLUDED.low_price,
			close_price = EXCLUDED.close_price,
			volume      = EXCLUDED.volume,
			amount      = EXCLUDED.amount`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, b := range bars {
		_, err := tx.ExecContext(ctx, stmt,
			strings.ToUpper(b.IndexCode), b.TradeDate, b.Open, b.High, b.Low, b.Close,
			b.Volume, b.Amount, b.DataVersion,
		)
		if err != nil {
			return fmt.Errorf("upsert index_bar %s %s: %w", b.IndexCode, b.TradeDate, err)
		}
	}
	return tx.Commit()
}
