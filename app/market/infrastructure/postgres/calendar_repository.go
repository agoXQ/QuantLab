package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// CalendarRepository persists trading calendar entries.
type CalendarRepository struct{ db *sql.DB }

// NewCalendarRepository creates a new CalendarRepository.
func NewCalendarRepository(db *sql.DB) *CalendarRepository { return &CalendarRepository{db: db} }

// Range implements calendar.Repository.
func (r *CalendarRepository) Range(ctx context.Context, rg valueobject.DateRange) ([]calendar.TradingDay, error) {
	args := []any{}
	conditions := []string{"1=1"}
	if !rg.Start.IsZero() {
		args = append(args, rg.Start)
		conditions = append(conditions, fmt.Sprintf("trade_date >= $%d", len(args)))
	}
	if !rg.End.IsZero() {
		args = append(args, rg.End)
		conditions = append(conditions, fmt.Sprintf("trade_date <= $%d", len(args)))
	}
	query := fmt.Sprintf(`SELECT trade_date, is_open FROM trading_calendar WHERE %s ORDER BY trade_date ASC`, joinAnd(conditions))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query trading_calendar: %w", err)
	}
	defer rows.Close()

	out := make([]calendar.TradingDay, 0, 256)
	for rows.Next() {
		var d calendar.TradingDay
		if err := rows.Scan(&d.TradeDate, &d.IsOpen); err != nil {
			return nil, fmt.Errorf("scan trading_calendar: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// BulkUpsert implements calendar.Repository.
func (r *CalendarRepository) BulkUpsert(ctx context.Context, days []calendar.TradingDay) error {
	if len(days) == 0 {
		return nil
	}
	const stmt = `
		INSERT INTO trading_calendar (trade_date, is_open) VALUES ($1, $2)
		ON CONFLICT (trade_date) DO UPDATE SET is_open = EXCLUDED.is_open`
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, d := range days {
		if _, err := tx.ExecContext(ctx, stmt, d.TradeDate, d.IsOpen); err != nil {
			return fmt.Errorf("upsert calendar %s: %w", d.TradeDate, err)
		}
	}
	return tx.Commit()
}

func joinAnd(parts []string) string {
	out := parts[0]
	for _, p := range parts[1:] {
		out += " AND " + p
	}
	return out
}
