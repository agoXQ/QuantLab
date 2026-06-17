// Package indexbar defines the IndexBar aggregate for index daily quotes.
package indexbar

import (
	"context"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// IndexBar represents a daily index OHLC snapshot.
type IndexBar struct {
	IndexCode   string    `json:"index_code"`
	TradeDate   time.Time `json:"trade_date"`
	Open        float64   `json:"open,omitempty"`
	High        float64   `json:"high,omitempty"`
	Low         float64   `json:"low,omitempty"`
	Close       float64   `json:"close"`
	Volume      int64     `json:"volume,omitempty"`
	Amount      float64   `json:"amount,omitempty"`
	DataVersion string    `json:"data_version"`
}

// Repository persists IndexBar entities.
type Repository interface {
	List(ctx context.Context, q RangeQuery) ([]*IndexBar, error)
	BulkUpsert(ctx context.Context, bars []*IndexBar) error
}

// RangeQuery describes a query against index storage.
type RangeQuery struct {
	IndexCode   string
	Range       valueobject.DateRange
	DataVersion string
	Limit       int
}
