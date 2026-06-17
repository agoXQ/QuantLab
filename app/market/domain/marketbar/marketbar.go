// Package marketbar defines the MarketBar aggregate (K-line) for the Market Data service.
package marketbar

import (
	"context"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// MarketBar represents a single OHLCV bar bound to a stock and trade date.
//
// It is an immutable record once persisted. Adjustments are applied on read
// based on the requested adjustment mode and the AdjFactor stored alongside
// the raw bar.
type MarketBar struct {
	StockCode   string             `json:"stock_code"`
	TradeDate   time.Time          `json:"trade_date"`
	Period      valueobject.Period `json:"period"`
	Open        float64            `json:"open"`
	High        float64            `json:"high"`
	Low         float64            `json:"low"`
	Close       float64            `json:"close"`
	Volume      int64              `json:"volume"`
	Amount      float64            `json:"amount"`
	AdjFactor   float64            `json:"adj_factor"`
	DataVersion string             `json:"data_version"`
}

// Repository persists MarketBar entities.
type Repository interface {
	// Range returns raw (unadjusted) bars in the inclusive range. Bars must be
	// sorted by trade_date ascending.
	Range(ctx context.Context, q RangeQuery) ([]*MarketBar, error)

	// Latest returns the most recent bar at or before the cutoff. Returns nil
	// when no bar exists.
	Latest(ctx context.Context, stockCode string, period valueobject.Period, cutoff time.Time) (*MarketBar, error)

	// BulkUpsert writes a batch of bars within a single transaction.
	BulkUpsert(ctx context.Context, bars []*MarketBar) error
}

// RangeQuery describes a query against MarketBar storage.
type RangeQuery struct {
	StockCode   string
	Period      valueobject.Period
	Range       valueobject.DateRange
	DataVersion string
	Limit       int
}

// Validate ensures the query has the required fields populated.
func (q RangeQuery) Validate() error {
	if q.StockCode == "" {
		return ErrEmptyStockCode
	}
	if !q.Period.IsValid() {
		return ErrEmptyStockCode // sentinel; concrete error left to callers via valueobject parsing
	}
	if err := q.Range.Validate(); err != nil {
		return err
	}
	return nil
}

// ErrEmptyStockCode is a domain-only sentinel; callers should translate to a
// MarketError before crossing layers. Keeping it private avoids leaking
// validation details from the domain.
var ErrEmptyStockCode = errEmpty("stock_code is required")

type errEmpty string

func (e errEmpty) Error() string { return string(e) }
