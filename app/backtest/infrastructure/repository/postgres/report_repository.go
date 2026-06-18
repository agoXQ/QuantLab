package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	"github.com/agoXQ/QuantLab/app/backtest/domain/report"
)

// ReportRepository persists PerformanceReport aggregates in PostgreSQL.
type ReportRepository struct {
	db *sql.DB
}

// NewReportRepository wires the repository.
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Save upserts the report on the job_id primary key so reruns overwrite
// stale numbers without leaving orphan rows.
func (r *ReportRepository) Save(ctx context.Context, rep *report.PerformanceReport) error {
	if rep == nil {
		return bterr.ErrReportNotFound
	}
	curve := rep.EquityCurve
	if curve == nil {
		curve = []report.EquityPoint{}
	}
	payload, err := json.Marshal(curve)
	if err != nil {
		return fmt.Errorf("marshal equity_curve: %w", err)
	}
	const stmt = `
		INSERT INTO backtest_report (
			job_id, start_date, end_date, initial_capital, final_asset,
			total_return, annual_return, volatility, sharpe_ratio,
			max_drawdown, win_rate, trade_count, equity_curve, generated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (job_id) DO UPDATE SET
			start_date      = EXCLUDED.start_date,
			end_date        = EXCLUDED.end_date,
			initial_capital = EXCLUDED.initial_capital,
			final_asset     = EXCLUDED.final_asset,
			total_return    = EXCLUDED.total_return,
			annual_return   = EXCLUDED.annual_return,
			volatility      = EXCLUDED.volatility,
			sharpe_ratio    = EXCLUDED.sharpe_ratio,
			max_drawdown    = EXCLUDED.max_drawdown,
			win_rate        = EXCLUDED.win_rate,
			trade_count     = EXCLUDED.trade_count,
			equity_curve    = EXCLUDED.equity_curve,
			generated_at    = EXCLUDED.generated_at`
	if _, err := r.db.ExecContext(ctx, stmt,
		rep.JobID, rep.StartDate, rep.EndDate,
		rep.InitialCapital, rep.FinalAsset,
		rep.TotalReturn, rep.AnnualReturn, rep.Volatility, rep.SharpeRatio,
		rep.MaxDrawdown, rep.WinRate, rep.TradeCount,
		payload, rep.GeneratedAt,
	); err != nil {
		return fmt.Errorf("upsert backtest_report: %w", err)
	}
	return nil
}

// Get loads the report for a job. Returns ErrReportNotFound when the row is
// absent so the application layer can surface a 404 cleanly.
func (r *ReportRepository) Get(ctx context.Context, jobID int64) (*report.PerformanceReport, error) {
	const q = `
		SELECT job_id, start_date, end_date, initial_capital, final_asset,
		       total_return, annual_return, volatility, sharpe_ratio,
		       max_drawdown, win_rate, trade_count, equity_curve, generated_at
		FROM backtest_report
		WHERE job_id = $1`
	row := r.db.QueryRowContext(ctx, q, jobID)
	var (
		rep     report.PerformanceReport
		payload []byte
	)
	err := row.Scan(
		&rep.JobID, &rep.StartDate, &rep.EndDate,
		&rep.InitialCapital, &rep.FinalAsset,
		&rep.TotalReturn, &rep.AnnualReturn, &rep.Volatility, &rep.SharpeRatio,
		&rep.MaxDrawdown, &rep.WinRate, &rep.TradeCount,
		&payload, &rep.GeneratedAt,
	)
	if err == sql.ErrNoRows {
		return nil, bterr.ErrReportNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan backtest_report: %w", err)
	}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &rep.EquityCurve); err != nil {
			return nil, fmt.Errorf("unmarshal equity_curve: %w", err)
		}
	}
	return &rep, nil
}
