// Package marketdata defines the read-side abstractions Backtest needs from
// Market Data. Concrete adapters live under infrastructure/marketdata and
// translate to either market.Service (in-process) or a future gRPC client.
//
// The interface is deliberately narrow: Backtest never asks for fundamentals
// or corporate actions, only for the daily OHLCV slice it needs to mark
// positions and the trading calendar that drives the replay loop.
package marketdata

import (
	"context"
	"time"

	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// Bar is the OHLCV record consumed by the engine. Adjusted prices are
// expected: providers must apply the configured adjustment mode before
// returning so the engine never re-adjusts mid-replay.
type Bar struct {
	StockCode string    `json:"stock_code"`
	TradeDate time.Time `json:"trade_date"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	Amount    float64   `json:"amount"`
}

// Provider is the read side the engine depends on.
type Provider interface {
	// LoadBars returns the bars in the inclusive [Start, End] window for
	// each requested stock. Missing stocks must be returned with an empty
	// slice rather than as an error so a single suspended security does
	// not poison the entire run.
	LoadBars(ctx context.Context, req BarsRequest) (map[string][]Bar, error)

	// TradingDays returns the trading calendar in the inclusive window.
	// Day-by-day replay walks this list, which guarantees the engine
	// never invents a trading day from a calendar gap.
	TradingDays(ctx context.Context, req CalendarRequest) ([]time.Time, error)
}

// BarsRequest is the batched K-line request shape.
type BarsRequest struct {
	StockCodes  []string
	Range       valueobject.DateRange
	DataVersion string
}

// CalendarRequest scopes the trading calendar lookup.
type CalendarRequest struct {
	Range  valueobject.DateRange
	Market string
}
