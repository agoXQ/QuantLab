// Package portfolio defines the Portfolio aggregate. A Portfolio holds the
// running cash balance, the per-stock positions, and the daily snapshots
// used to compute performance metrics. Trades are the primary mutator;
// daily mark-to-market is performed by the engine before a snapshot is
// committed.
package portfolio

import (
	"context"
	"math"
	"time"

	"github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// Position represents a long-only holding of one stock.
type Position struct {
	StockCode    string  `json:"stock_code"`
	Quantity     int64   `json:"quantity"`
	CostPrice    float64 `json:"cost_price"`
	MarketPrice  float64 `json:"market_price"`
	MarketValue  float64 `json:"market_value"`
}

// Snapshot is the per-trade-date cross-section of the portfolio. The engine
// emits one snapshot at the close of every replay tick.
type Snapshot struct {
	JobID       int64      `json:"job_id"`
	TradeDate   time.Time  `json:"trade_date"`
	Cash        float64    `json:"cash"`
	MarketValue float64    `json:"market_value"`
	TotalAsset  float64    `json:"total_asset"`
	Positions   []Position `json:"positions"`
}

// Portfolio is the in-memory aggregate the engine mutates while replaying.
//
// The persistence side stores Snapshots and Trades; Portfolio itself is
// rebuilt from those tables when a job resumes. The aggregate intentionally
// keeps no event log of its own to avoid double-bookkeeping with Trade.
type Portfolio struct {
	JobID     int64
	Cash      float64
	Positions map[string]*Position
}

// New creates an empty Portfolio with the given starting cash.
func New(jobID int64, initialCash float64) *Portfolio {
	return &Portfolio{
		JobID:     jobID,
		Cash:      initialCash,
		Positions: make(map[string]*Position),
	}
}

// Position returns the position for stockCode or nil when none is held.
func (p *Portfolio) Position(stockCode string) *Position {
	if pos, ok := p.Positions[stockCode]; ok {
		return pos
	}
	return nil
}

// ApplyTrade folds a Trade into the cash balance and the relevant Position.
// The math is simple long-only weighted average cost; shorting and option
// legs are out of scope for the MVP.
func (p *Portfolio) ApplyTrade(t *trade.Trade) {
	if t == nil || t.Quantity == 0 {
		return
	}
	switch t.Side {
	case valueobject.OrderSideBuy:
		p.Cash -= t.NetAmount()
		pos, ok := p.Positions[t.StockCode]
		if !ok {
			p.Positions[t.StockCode] = &Position{
				StockCode:   t.StockCode,
				Quantity:    t.Quantity,
				CostPrice:   t.Price,
				MarketPrice: t.Price,
				MarketValue: float64(t.Quantity) * t.Price,
			}
			return
		}
		newQty := pos.Quantity + t.Quantity
		if newQty <= 0 {
			delete(p.Positions, t.StockCode)
			return
		}
		// Average-cost update including fees so post-trade PnL reflects
		// the all-in entry price the user actually paid.
		totalCost := pos.CostPrice*float64(pos.Quantity) + t.Price*float64(t.Quantity) + t.Commission + t.Slippage
		pos.Quantity = newQty
		pos.CostPrice = totalCost / float64(newQty)
		pos.MarketPrice = t.Price
		pos.MarketValue = float64(newQty) * t.Price
	case valueobject.OrderSideSell:
		// SELL produces positive cash inflow (NetAmount is negative).
		p.Cash -= t.NetAmount()
		pos, ok := p.Positions[t.StockCode]
		if !ok {
			return
		}
		pos.Quantity -= t.Quantity
		if pos.Quantity <= 0 {
			delete(p.Positions, t.StockCode)
			return
		}
		pos.MarketPrice = t.Price
		pos.MarketValue = float64(pos.Quantity) * t.Price
	}
}

// MarkToMarket updates per-position MarketPrice/MarketValue using a price
// map keyed by stock code. Missing prices leave the existing mark intact;
// callers can decide whether to surface that as a halted security.
func (p *Portfolio) MarkToMarket(prices map[string]float64) {
	for code, pos := range p.Positions {
		if px, ok := prices[code]; ok && !math.IsNaN(px) {
			pos.MarketPrice = px
			pos.MarketValue = float64(pos.Quantity) * px
		}
	}
}

// Snapshot returns a Snapshot for the given trade date, copying the current
// positions so callers may persist them safely.
func (p *Portfolio) Snapshot(tradeDate time.Time) Snapshot {
	positions := make([]Position, 0, len(p.Positions))
	market := 0.0
	for _, pos := range p.Positions {
		positions = append(positions, *pos)
		market += pos.MarketValue
	}
	return Snapshot{
		JobID:       p.JobID,
		TradeDate:   tradeDate,
		Cash:        p.Cash,
		MarketValue: market,
		TotalAsset:  p.Cash + market,
		Positions:   positions,
	}
}

// Repository persists Portfolio Snapshots scoped to a single BacktestJob.
type Repository interface {
	BulkInsertSnapshots(ctx context.Context, snapshots []Snapshot) error
	ListSnapshots(ctx context.Context, jobID int64) ([]Snapshot, error)
}
