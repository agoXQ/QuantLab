// Package executor defines the StrategyExecutor port: the bridge between a
// formula plan and a list of buy / sell signals at a given trade date.
//
// Backtest does not embed the Formula Engine directly; it talks to whatever
// adapter is wired in (in-process EvaluatorService for the monolith mode,
// gRPC client later). This keeps the engine free of compiler / planner
// imports and makes the Formula <-> Backtest seam explicit.
package executor

import (
	"context"
	"time"
)

// SignalAction describes the desired position change for a stock.
type SignalAction string

const (
	SignalBuy  SignalAction = "BUY"
	SignalSell SignalAction = "SELL"
	SignalHold SignalAction = "HOLD"
)

// Signal is a single per-stock instruction emitted by the strategy.
//
// Score is optional: it is the ranking score for SORT plans and the raw
// VALUE for VALUE plans. The matching engine ignores Score; the order
// generator uses it for weight assignment.
type Signal struct {
	StockCode string
	Action    SignalAction
	Score     float64
}

// Request bundles the inputs required to produce signals at one trade date.
type Request struct {
	Formula      string
	Universe     []string
	AsOfDate     time.Time
	LookbackBars int
	DataVersion  string
}

// Result is the discriminated outcome of a single execution. Selection is
// the canonical input for FILTER plans; Ranking carries scores so the
// portfolio engine can build weights; Values surfaces raw VALUE plans for
// signals like "buy when MACD crosses zero".
type Result struct {
	Signals []Signal
}

// StrategyExecutor turns a formula + universe into Signals.
type StrategyExecutor interface {
	Execute(ctx context.Context, req Request) (*Result, error)
}
