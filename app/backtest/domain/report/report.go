// Package report defines the PerformanceReport aggregate.
//
// A report is the canonical summary of a finished BacktestJob: the
// headline metrics (annualised return, max drawdown, Sharpe, win rate)
// and the daily equity curve frontends use to draw charts. The drawdown
// series is intentionally derived from the equity curve so callers do not
// have to recompute it client-side.
package report

import (
	"context"
	"time"
)

// EquityPoint is a single point on the equity curve.
type EquityPoint struct {
	TradeDate  time.Time `json:"trade_date"`
	TotalAsset float64   `json:"total_asset"`
	Drawdown   float64   `json:"drawdown"`
	Return     float64   `json:"return"`
}

// PerformanceReport bundles the headline metrics and supporting series.
type PerformanceReport struct {
	JobID         int64         `json:"job_id"`
	StartDate     time.Time     `json:"start_date"`
	EndDate       time.Time     `json:"end_date"`
	InitialCapital float64      `json:"initial_capital"`
	FinalAsset    float64       `json:"final_asset"`
	TotalReturn   float64       `json:"total_return"`
	AnnualReturn  float64       `json:"annual_return"`
	Volatility    float64       `json:"volatility"`
	SharpeRatio   float64       `json:"sharpe_ratio"`
	MaxDrawdown   float64       `json:"max_drawdown"`
	WinRate       float64       `json:"win_rate"`
	TradeCount    int           `json:"trade_count"`
	EquityCurve   []EquityPoint `json:"equity_curve"`
	GeneratedAt   time.Time     `json:"generated_at"`
}

// Repository persists PerformanceReport aggregates.
type Repository interface {
	Save(ctx context.Context, r *PerformanceReport) error
	Get(ctx context.Context, jobID int64) (*PerformanceReport, error)
}
