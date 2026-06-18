// Package trade defines the Trade entity. A Trade is the immutable record
// of a filled Order; it is the canonical input for performance analysis.
package trade

import (
	"context"
	"time"

	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// Trade represents a single fill against an Order.
type Trade struct {
	ID         int64                  `json:"id"`
	JobID      int64                  `json:"job_id"`
	OrderID    int64                  `json:"order_id"`
	StockCode  string                 `json:"stock_code"`
	Side       valueobject.OrderSide  `json:"side"`
	Quantity   int64                  `json:"quantity"`
	Price      float64                `json:"price"`
	Commission float64                `json:"commission"`
	StampDuty  float64                `json:"stamp_duty"`
	Slippage   float64                `json:"slippage"`
	TradeTime  time.Time              `json:"trade_time"`
}

// GrossAmount returns Quantity * Price.
func (t Trade) GrossAmount() float64 { return float64(t.Quantity) * t.Price }

// NetAmount returns the cash impact of the trade. Positive means cash leaves
// the portfolio (BUY); negative means cash flows in (SELL after fees).
func (t Trade) NetAmount() float64 {
	gross := t.GrossAmount()
	fees := t.Commission + t.StampDuty
	switch t.Side {
	case valueobject.OrderSideBuy:
		return gross + fees
	case valueobject.OrderSideSell:
		return -(gross - fees)
	default:
		return 0
	}
}

// Repository persists Trade entities scoped to a single BacktestJob.
type Repository interface {
	BulkInsert(ctx context.Context, trades []*Trade) error
	ListByJob(ctx context.Context, jobID int64) ([]*Trade, error)
}
