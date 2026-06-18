// Package matching defines the order-matching contract.
//
// MVP semantics: orders submitted at the close of day D are matched at the
// open of day D+1 with a configurable slippage band (T+1 next-open). The
// engine never revisits past dates, so the matching call is pure with
// respect to the BarSnapshot it receives.
package matching

import (
	"github.com/agoXQ/QuantLab/app/backtest/domain/order"
	"github.com/agoXQ/QuantLab/app/backtest/domain/trade"
)

// BarSnapshot bundles the bar info the matcher needs to fill a single order.
//
// Halted is true when the security cannot be traded on this bar (suspended,
// pre-listing, etc.). LimitUp / LimitDown gate buy / sell orders to enforce
// A-share daily price limits.
type BarSnapshot struct {
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
	Halted    bool
	LimitUp   bool
	LimitDown bool
}

// Engine matches a batch of orders against per-stock BarSnapshots.
//
// The map is indexed by stock code. Orders for missing or halted stocks
// must be returned as rejected orders (ord.Status == REJECTED) rather
// than dropped silently so the API trades view can surface them.
type Engine interface {
	Match(orders []*order.Order, bars map[string]BarSnapshot) ([]*trade.Trade, []*order.Order)
}
