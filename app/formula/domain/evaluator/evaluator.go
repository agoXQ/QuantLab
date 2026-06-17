// Package evaluator defines the contract for evaluating a compiled formula
// against market data.
//
// The evaluator turns an ExecutionPlan into one of three shapes:
//
//   FILTER  -> Selection of stock codes (used for stock screening)
//   SORT    -> ranked list of (stock, score) pairs (used for ranking strategies)
//   VALUE   -> per-stock numeric or boolean value (used for ad-hoc inspection)
//
// Implementations must be safe for concurrent use; the package keeps no
// shared state of its own. The data dependency is satisfied through the
// DataPort interface defined here, which adapters implement against the
// Market Data Service.
package evaluator

import (
	"context"
	"time"

	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	domainSeries "github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// Evaluator runs a compiled execution plan.
type Evaluator interface {
	// Evaluate produces a Result for the given plan and request. The plan
	// must have been produced by compiler.Planner; the evaluator does not
	// re-run validation. Implementations must respect ctx cancellation.
	Evaluate(ctx context.Context, plan *domainCompiler.ExecutionPlan, req Request) (*Result, error)
}

// Request bundles the runtime inputs an evaluator needs.
//
// Universe is the set of stocks the formula is evaluated against. AsOfDate is
// the cross-section date at which the formula is materialised; for time-series
// FILTERs this corresponds to the bar whose value drives the inclusion test.
// LookbackBars caps the historical window the evaluator may request from the
// DataPort; it must be large enough to satisfy every indicator inside the
// plan (e.g. MA(CLOSE, 60) requires LookbackBars >= 60).
type Request struct {
	Universe     []string
	AsOfDate     time.Time
	LookbackBars int
	DataVersion  string
}

// Result is the discriminated outcome of an evaluation.
//
// PlanType mirrors the plan input so callers can branch without re-inspecting
// the AST.
type Result struct {
	PlanType  domainCompiler.PlanType
	Selection *Selection
	Ranking   *Ranking
	Values    *ValueMap
}

// Selection is the set of stocks that satisfy a FILTER plan.
//
// Codes are returned in the order they appear in the input universe so that
// downstream consumers (e.g. the backtest engine) can reproduce results
// deterministically.
type Selection struct {
	StockCodes []string
}

// Ranking is the result of a SORT plan.
//
// Scores are aligned with StockCodes; NaN scores are dropped before sorting.
type Ranking struct {
	StockCodes []string
	Scores     []float64
}

// ValueMap is the result of a VALUE plan: one final value per stock. A
// trailing-NaN value indicates the formula produced no data for that stock
// (insufficient history, missing financials, etc.).
type ValueMap struct {
	StockCodes []string
	Values     []float64
}

// DataPort is the read side that the evaluator depends on.
//
// Adapters implement this interface against the Market Data Service. The
// evaluator never touches a database or RPC client directly.
type DataPort interface {
	// LoadBars returns up to LookbackBars + 1 bars per stock, ending at
	// AsOfDate (inclusive). Bars must be sorted by timestamp ascending; the
	// evaluator does not re-sort them. Missing stocks must be returned with
	// an empty bar slice rather than as a hard error so a single missing
	// security does not poison a large universe evaluation.
	LoadBars(ctx context.Context, req BarsRequest) (map[string][]domainSeries.Bar, error)

	// LoadFinancialsLatest returns the latest known financial metric for each
	// stock, as observed on or before AsOfDate. The map key is the stock
	// code; the inner map key is the canonical metric name (PE, PB, ROE,
	// ROA, EPS, RevenueGrowth, ProfitGrowth, MarketCap, FloatMarketCap).
	// Missing values must be omitted from the inner map; the evaluator
	// substitutes NaN.
	LoadFinancialsLatest(ctx context.Context, req FinancialsRequest) (map[string]map[string]float64, error)
}

// BarsRequest describes a batched K-line request.
type BarsRequest struct {
	StockCodes   []string
	AsOfDate     time.Time
	LookbackBars int
	DataVersion  string
}

// FinancialsRequest describes a batched financial metric request.
type FinancialsRequest struct {
	StockCodes  []string
	Metrics     []string
	AsOfDate    time.Time
	DataVersion string
}
