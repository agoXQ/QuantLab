package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// MarketBarRepository persists MarketBar entities in PostgreSQL/TimescaleDB.
type MarketBarRepository struct {
	db *sql.DB
}

// NewMarketBarRepository creates a new MarketBarRepository.
func NewMarketBarRepository(db *sql.DB) *MarketBarRepository {
	return &MarketBarRepository{db: db}
}

// Range implements marketbar.Repository.
func (r *MarketBarRepository) Range(ctx context.Context, q marketbar.RangeQuery) ([]*marketbar.MarketBar, error) {
	if strings.TrimSpace(q.StockCode) == "" {
		return nil, marketbar.ErrEmptyStockCode
	}
	period := q.Period
	if period == "" {
		period = valueobject.PeriodDay
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 1000
	}

	args := []any{strings.ToUpper(q.StockCode), string(period)}
	conditions := []string{"stock_code = $1", "period = $2"}
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
	orderDirection := "ASC"
	wrapAscending := false
	if q.Range.Start.IsZero() && !q.Range.End.IsZero() {
		orderDirection = "DESC"
		wrapAscending = true
	}

	baseQuery := fmt.Sprintf(`
		SELECT stock_code, trade_date, period,
		       COALESCE(open_price, 0), COALESCE(high_price, 0),
		       COALESCE(low_price, 0), COALESCE(close_price, 0),
		       COALESCE(volume, 0), COALESCE(amount, 0),
		       COALESCE(adj_factor, 1), COALESCE(data_version, '')
		FROM market_bar
		WHERE %s
		ORDER BY trade_date %s
		LIMIT $%d`, strings.Join(conditions, " AND "), orderDirection, len(args))
	query := baseQuery
	if wrapAscending {
		query = fmt.Sprintf("SELECT * FROM (%s) recent_bars ORDER BY trade_date ASC", baseQuery)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query market_bar: %w", err)
	}
	defer rows.Close()

	out := make([]*marketbar.MarketBar, 0, limit)
	for rows.Next() {
		var (
			b      marketbar.MarketBar
			period string
		)
		if err := rows.Scan(
			&b.StockCode, &b.TradeDate, &period,
			&b.Open, &b.High, &b.Low, &b.Close,
			&b.Volume, &b.Amount, &b.AdjFactor, &b.DataVersion,
		); err != nil {
			return nil, fmt.Errorf("scan market_bar: %w", err)
		}
		b.Period = valueobject.Period(period)
		out = append(out, &b)
	}
	return out, rows.Err()
}

// Latest implements marketbar.Repository.
func (r *MarketBarRepository) Latest(ctx context.Context, stockCode string, period valueobject.Period, cutoff time.Time) (*marketbar.MarketBar, error) {
	const q = `
		SELECT stock_code, trade_date, period,
		       COALESCE(open_price, 0), COALESCE(high_price, 0),
		       COALESCE(low_price, 0), COALESCE(close_price, 0),
		       COALESCE(volume, 0), COALESCE(amount, 0),
		       COALESCE(adj_factor, 1), COALESCE(data_version, '')
		FROM market_bar
		WHERE stock_code = $1 AND period = $2 AND trade_date <= $3
		ORDER BY trade_date DESC
		LIMIT 1`
	if cutoff.IsZero() {
		cutoff = time.Now().UTC()
	}
	row := r.db.QueryRowContext(ctx, q, strings.ToUpper(stockCode), string(period), cutoff)
	var (
		b   marketbar.MarketBar
		per string
	)
	err := row.Scan(
		&b.StockCode, &b.TradeDate, &per,
		&b.Open, &b.High, &b.Low, &b.Close,
		&b.Volume, &b.Amount, &b.AdjFactor, &b.DataVersion,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b.Period = valueobject.Period(per)
	return &b, nil
}

// BulkUpsert implements marketbar.Repository.
func (r *MarketBarRepository) BulkUpsert(ctx context.Context, bars []*marketbar.MarketBar) error {
	if len(bars) == 0 {
		return nil
	}
	const stmt = `
		INSERT INTO market_bar (
			stock_code, trade_date, period,
			open_price, high_price, low_price, close_price,
			volume, amount, adj_factor, data_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (stock_code, trade_date, period, data_version) DO UPDATE SET
			open_price  = EXCLUDED.open_price,
			high_price  = EXCLUDED.high_price,
			low_price   = EXCLUDED.low_price,
			close_price = EXCLUDED.close_price,
			volume      = EXCLUDED.volume,
			amount      = EXCLUDED.amount,
			adj_factor  = EXCLUDED.adj_factor`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, b := range bars {
		_, err := tx.ExecContext(ctx, stmt,
			strings.ToUpper(b.StockCode), b.TradeDate, string(b.Period),
			b.Open, b.High, b.Low, b.Close,
			b.Volume, b.Amount, b.AdjFactor, b.DataVersion,
		)
		if err != nil {
			return fmt.Errorf("upsert market_bar %s %s: %w", b.StockCode, b.TradeDate, err)
		}
	}
	return tx.Commit()
}
