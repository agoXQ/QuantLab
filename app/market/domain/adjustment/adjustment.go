// Package adjustment defines the price adjustment service interface.
//
// Splits, bonus shares, and dividends introduce discontinuities in raw price
// series. The adjustment service applies the requested adjustment policy to a
// slice of raw bars and returns adjusted bars with stable prices for charting,
// indicator computation, and backtesting.
package adjustment

import (
	"context"

	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// CorporateAction is a domain-side view of a single corporate event used to
// derive adjustment factors.
type CorporateAction struct {
	StockCode  string                          `json:"stock_code"`
	ActionDate string                          `json:"action_date"`
	ActionType valueobject.CorporateActionType `json:"action_type"`
	Factor     float64                         `json:"factor"`
}

// Adjuster applies an adjustment policy to a slice of raw bars.
//
// Implementations must not mutate the input slice; they should return a new
// slice with adjusted prices and an updated AdjFactor.
type Adjuster interface {
	Apply(ctx context.Context, bars []*marketbar.MarketBar, mode valueobject.Adjustment) ([]*marketbar.MarketBar, error)
}
