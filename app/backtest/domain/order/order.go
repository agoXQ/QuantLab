// Package order defines the Order entity inside the BacktestJob aggregate
// boundary. Orders are short-lived: they are emitted by the rebalance
// step, matched by the matching engine on the next trade date, and either
// turn into a Trade or get rejected (suspended / limit / insufficient
// cash). The aggregate root is BacktestJob; this package only models the
// shape and provides the repository contract for persistence.
package order

import (
	"context"
	"time"

	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// Order is a single buy/sell instruction queued for matching.
type Order struct {
	ID          int64                  `json:"id"`
	JobID       int64                  `json:"job_id"`
	StockCode   string                 `json:"stock_code"`
	Side        valueobject.OrderSide  `json:"side"`
	Quantity    int64                  `json:"quantity"`
	LimitPrice  float64                `json:"limit_price,omitempty"`
	Status      valueobject.OrderStatus `json:"status"`
	Reason      string                 `json:"reason,omitempty"`
	SubmittedAt time.Time              `json:"submitted_at"`
	FilledAt    *time.Time             `json:"filled_at,omitempty"`
	FilledPrice float64                `json:"filled_price,omitempty"`
	FilledQty   int64                  `json:"filled_qty,omitempty"`
}

// Repository persists Order entities scoped to a single BacktestJob.
type Repository interface {
	BulkInsert(ctx context.Context, orders []*Order) error
	ListByJob(ctx context.Context, jobID int64) ([]*Order, error)
}
