package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	btvo "github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
	marketVO "github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// Scenario is the on-disk schema for a backtest_e2e run. It mirrors the
// application's CreateBacktestRequest plus an `ingest` block that drives
// the Market Data ingestion service.
//
// The struct is annotated for yaml.v3; JSON tags are added on the side
// because the baseline file the harness writes uses JSON for diff
// friendliness.
type Scenario struct {
	Name        string         `yaml:"name"        json:"name"`
	Description string         `yaml:"description" json:"description,omitempty"`
	Ingest      IngestSpec     `yaml:"ingest"      json:"ingest"`
	Backtest    BacktestSpec   `yaml:"backtest"    json:"backtest"`
	Baseline    BaselineConfig `yaml:"baseline"    json:"baseline"`
}

// IngestSpec configures the Tushare ingestion stage. Each sub-block is
// optional; an empty block means "skip this fetch".
type IngestSpec struct {
	Skip        bool          `yaml:"skip"`
	Market      string        `yaml:"market"`
	Description string        `yaml:"description"`
	Bars        *BarsSpec     `yaml:"bars,omitempty"`
	Calendar    *CalendarSpec `yaml:"calendar,omitempty"`
	Financials  *FinSpec      `yaml:"financials,omitempty"`
	Factors     *FactorSpec   `yaml:"factors,omitempty"`
}

// BarsSpec is a daily-bar fetch over a date range.
type BarsSpec struct {
	Range  RangeSpec `yaml:"range"`
	Period string    `yaml:"period"`
}

// CalendarSpec captures a calendar fetch.
type CalendarSpec struct {
	Range RangeSpec `yaml:"range"`
}

// FinSpec captures a financial-statement fetch.
type FinSpec struct {
	Range      RangeSpec `yaml:"range"`
	ReportType string    `yaml:"report_type"`
}

// FactorSpec captures a factor fetch.
type FactorSpec struct {
	Range RangeSpec `yaml:"range"`
	Names []string  `yaml:"names"`
}

// BacktestSpec captures the backtest job configuration. It is a 1:1
// mirror of appBacktest.CreateBacktestRequest with shorter field names.
type BacktestSpec struct {
	Formula        string             `yaml:"formula"`
	Universe       []string           `yaml:"universe"`
	DataVersion    string             `yaml:"data_version"`
	InitialCapital float64            `yaml:"initial_capital"`
	Benchmark      string             `yaml:"benchmark"`
	Range          RangeSpec          `yaml:"range"`
	Config         BacktestConfigSpec `yaml:"config"`
}

// BacktestConfigSpec is the subset of backtestjob.Config a scenario can
// override. Fields left at zero use the application-side defaults.
type BacktestConfigSpec struct {
	RebalanceFrequency string  `yaml:"rebalance_frequency"`
	MaxPositionCount   int     `yaml:"max_position_count"`
	LookbackBars       int     `yaml:"lookback_bars"`
	CommissionRate     float64 `yaml:"commission_rate"`
	SlippageRate       float64 `yaml:"slippage_rate"`
	StampDutyRate      float64 `yaml:"stamp_duty_rate"`
	MinCommission      float64 `yaml:"min_commission"`
}

// BaselineConfig configures regression diffing.
type BaselineConfig struct {
	Tolerance map[string]float64 `yaml:"tolerance"`
}

// RangeSpec is a YYYY-MM-DD inclusive date window.
type RangeSpec struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

// LoadScenario reads and validates a scenario file.
//
// The harness fails fast on schema problems so a typo never silently
// promotes to "run with defaults"; CI must see the error.
func LoadScenario(path string) (*Scenario, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario: %w", err)
	}
	var sc Scenario
	if err := yaml.Unmarshal(raw, &sc); err != nil {
		return nil, fmt.Errorf("parse scenario: %w", err)
	}
	if strings.TrimSpace(sc.Name) == "" {
		sc.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if err := sc.validate(); err != nil {
		return nil, err
	}
	return &sc, nil
}

func (s *Scenario) validate() error {
	if strings.TrimSpace(s.Backtest.Formula) == "" {
		return fmt.Errorf("backtest.formula is required")
	}
	if len(s.Backtest.Universe) == 0 {
		return fmt.Errorf("backtest.universe is required")
	}
	if s.Backtest.InitialCapital <= 0 {
		return fmt.Errorf("backtest.initial_capital must be positive")
	}
	if _, err := s.Backtest.Range.Parse(); err != nil {
		return fmt.Errorf("backtest.range: %w", err)
	}
	if !s.Ingest.Skip {
		if strings.TrimSpace(s.Ingest.Market) == "" {
			return fmt.Errorf("ingest.market is required when ingest is not skipped")
		}
	}
	return nil
}

// Parse turns the date pair into a marketVO.DateRange.
func (r RangeSpec) Parse() (marketVO.DateRange, error) {
	if strings.TrimSpace(r.Start) == "" || strings.TrimSpace(r.End) == "" {
		return marketVO.DateRange{}, fmt.Errorf("range start/end must be set")
	}
	start, err := marketVO.ParseDate(r.Start)
	if err != nil {
		return marketVO.DateRange{}, fmt.Errorf("parse start %q: %w", r.Start, err)
	}
	end, err := marketVO.ParseDate(r.End)
	if err != nil {
		return marketVO.DateRange{}, fmt.Errorf("parse end %q: %w", r.End, err)
	}
	if end.Before(start) {
		return marketVO.DateRange{}, fmt.Errorf("range end before start")
	}
	return marketVO.DateRange{Start: start, End: end}, nil
}

// BacktestRange materialises the scenario's backtest range into the
// matching backtest valueobject.DateRange.
func (s *Scenario) BacktestRange() (btvo.DateRange, error) {
	r, err := s.Backtest.Range.Parse()
	if err != nil {
		return btvo.DateRange{}, err
	}
	return btvo.DateRange{Start: r.Start, End: r.End}, nil
}

// BacktestConfig fills in a backtestjob.Config from the scenario.
func (s *Scenario) BacktestConfig() (backtestjob.Config, error) {
	cfg := backtestjob.Config{
		MaxPositionCount: s.Backtest.Config.MaxPositionCount,
		LookbackBars:     s.Backtest.Config.LookbackBars,
		CommissionRate:   s.Backtest.Config.CommissionRate,
		SlippageRate:     s.Backtest.Config.SlippageRate,
		StampDutyRate:    s.Backtest.Config.StampDutyRate,
		MinCommission:    s.Backtest.Config.MinCommission,
	}
	if rf := strings.TrimSpace(s.Backtest.Config.RebalanceFrequency); rf != "" {
		freq, err := btvo.ParseRebalanceFrequency(rf)
		if err != nil {
			return cfg, fmt.Errorf("parse rebalance_frequency: %w", err)
		}
		cfg.RebalanceFrequency = freq
	}
	cfg.Normalize()
	return cfg, nil
}

// IngestMarket parses the configured market identifier.
func (s *Scenario) IngestMarket() (marketVO.Market, error) {
	m := marketVO.Market(strings.ToUpper(strings.TrimSpace(s.Ingest.Market)))
	switch m {
	case marketVO.MarketCN, marketVO.MarketHK, marketVO.MarketUS:
		return m, nil
	default:
		return "", fmt.Errorf("unknown market %q", s.Ingest.Market)
	}
}

// BaselinePath returns the JSON baseline path for the scenario file.
func BaselinePath(scenarioFile string) string {
	dir := filepath.Dir(scenarioFile)
	base := strings.TrimSuffix(filepath.Base(scenarioFile), filepath.Ext(scenarioFile))
	return filepath.Join(dir, base+".baseline.json")
}

// Now is exported so tests / replay tools can override the wall clock.
var Now = func() time.Time { return time.Now().UTC() }
