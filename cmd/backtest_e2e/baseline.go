package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/agoXQ/QuantLab/app/backtest/domain/report"
)

// metricsSnapshot is the on-disk shape of a baseline. We pin only the
// fields that have well-defined diff semantics; equity curves and trade
// listings drift between runs (different fill timestamps) and would
// generate noisy false negatives.
type metricsSnapshot struct {
	GeneratedAt    time.Time `json:"generated_at"`
	Universe       []string  `json:"universe"`
	Formula        string    `json:"formula"`
	StartDate      string    `json:"start_date"`
	EndDate        string    `json:"end_date"`
	InitialCapital float64   `json:"initial_capital"`
	FinalAsset     float64   `json:"final_asset"`
	TotalReturn    float64   `json:"total_return"`
	AnnualReturn   float64   `json:"annual_return"`
	Volatility     float64   `json:"volatility"`
	SharpeRatio    float64   `json:"sharpe_ratio"`
	MaxDrawdown    float64   `json:"max_drawdown"`
	WinRate        float64   `json:"win_rate"`
	TradeCount     int       `json:"trade_count"`
}

// snapshotFrom collapses a PerformanceReport into a baseline-friendly
// payload, attaching the scenario surface so future readers can tell at
// a glance which run produced the file.
func snapshotFrom(rep *report.PerformanceReport, sc *Scenario) metricsSnapshot {
	universe := make([]string, len(sc.Backtest.Universe))
	copy(universe, sc.Backtest.Universe)
	sort.Strings(universe)
	return metricsSnapshot{
		GeneratedAt:    Now(),
		Universe:       universe,
		Formula:        sc.Backtest.Formula,
		StartDate:      formatDate(rep.StartDate),
		EndDate:        formatDate(rep.EndDate),
		InitialCapital: rep.InitialCapital,
		FinalAsset:     rep.FinalAsset,
		TotalReturn:    rep.TotalReturn,
		AnnualReturn:   rep.AnnualReturn,
		Volatility:     rep.Volatility,
		SharpeRatio:    rep.SharpeRatio,
		MaxDrawdown:    rep.MaxDrawdown,
		WinRate:        rep.WinRate,
		TradeCount:     rep.TradeCount,
	}
}

// readBaseline loads a previously written snapshot. A missing file is
// not an error; the caller treats it as "no baseline yet" and writes a
// fresh one.
func readBaseline(path string) (*metricsSnapshot, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read baseline: %w", err)
	}
	var ms metricsSnapshot
	if err := json.Unmarshal(data, &ms); err != nil {
		return nil, fmt.Errorf("parse baseline: %w", err)
	}
	return &ms, nil
}

// writeBaseline persists a snapshot in pretty JSON so reviewers can read
// the diff in a pull request without tooling.
func writeBaseline(path string, ms *metricsSnapshot) error {
	data, err := json.MarshalIndent(ms, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write baseline: %w", err)
	}
	return nil
}

// diff is one row of the comparison table the harness prints.
type diff struct {
	field     string
	baseline  float64
	current   float64
	tolerance float64
	pass      bool
}

// compareMetrics walks the metric fields, applies per-field tolerance,
// and reports whether the run is within bounds. Tolerance values come
// from scenario.Baseline.Tolerance with sensible fallbacks for fields
// the user did not configure.
func compareMetrics(prev, curr *metricsSnapshot, tol map[string]float64) (passed bool, rows []diff) {
	if tol == nil {
		tol = map[string]float64{}
	}
	check := func(name string, a, b float64) {
		t, ok := tol[name]
		if !ok {
			t = defaultTolerance(name)
		}
		row := diff{field: name, baseline: a, current: b, tolerance: t, pass: math.Abs(a-b) <= t}
		rows = append(rows, row)
		if !row.pass {
			passed = false
		}
	}
	passed = true

	check("total_return", prev.TotalReturn, curr.TotalReturn)
	check("annual_return", prev.AnnualReturn, curr.AnnualReturn)
	check("volatility", prev.Volatility, curr.Volatility)
	check("sharpe_ratio", prev.SharpeRatio, curr.SharpeRatio)
	check("max_drawdown", prev.MaxDrawdown, curr.MaxDrawdown)
	check("win_rate", prev.WinRate, curr.WinRate)
	check("final_asset", prev.FinalAsset, curr.FinalAsset)
	check("trade_count", float64(prev.TradeCount), float64(curr.TradeCount))
	return passed, rows
}

func defaultTolerance(name string) float64 {
	switch name {
	case "trade_count":
		return 2
	case "final_asset":
		// Loosely tied to total_return; let total_return guard the actual fail.
		return 1e6
	case "sharpe_ratio":
		return 0.5
	default:
		return 0.05
	}
}

func formatDate(t time.Time) string { return t.UTC().Format("2006-01-02") }
