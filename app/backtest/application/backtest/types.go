package backtest

import (
	"time"

	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	"github.com/agoXQ/QuantLab/app/backtest/domain/order"
	"github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
	"github.com/agoXQ/QuantLab/app/backtest/domain/report"
	"github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// CreateBacktestRequest is the payload for the Create use case.
//
// All fields except Formula / Universe / Range / InitialCapital are
// optional. Config is normalised to A-share defaults inside the service so
// callers can omit fields they do not care about.
type CreateBacktestRequest struct {
	UserID         int64
	StrategyID     int64
	VersionID      int64
	Name           string
	Formula        string
	Universe       []string
	Benchmark      string
	DataVersion    string
	InitialCapital float64
	Range          valueobject.DateRange
	Config         backtestjob.Config
}

// CreateBacktestResult is what the Create use case returns.
type CreateBacktestResult struct {
	Job *backtestjob.BacktestJob
}

// RunResult bundles the artefacts produced by a single synchronous run.
type RunResult struct {
	Job       *backtestjob.BacktestJob
	Report    *report.PerformanceReport
	Trades    []*trade.Trade
	Orders    []*order.Order
	Snapshots []portfolio.Snapshot
}

// ListJobsQuery is used by the API list endpoint.
type ListJobsQuery struct {
	UserID     int64
	StrategyID int64
	Status     valueobject.JobStatus
	Limit      int
}

// JobView packages a job alongside its metadata for the list response.
type JobView struct {
	Job *backtestjob.BacktestJob
}

// withTimestamp is a tiny helper used by callers that need a pointer to now.
func withTimestamp(t time.Time) *time.Time {
	cp := t
	return &cp
}

var _ = withTimestamp
