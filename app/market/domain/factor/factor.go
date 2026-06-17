// Package factor defines the Factor aggregate.
package factor

import (
	"context"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// Factor represents a derived factor value for a security on a trade date.
type Factor struct {
	StockCode   string    `json:"stock_code"`
	TradeDate   time.Time `json:"trade_date"`
	FactorName  string    `json:"factor_name"`
	FactorValue float64   `json:"factor_value"`
	DataVersion string    `json:"data_version"`
}

// Repository persists Factor entities.
type Repository interface {
	List(ctx context.Context, q ListQuery) ([]*Factor, error)
	BulkUpsert(ctx context.Context, factors []*Factor) error
}

// ListQuery describes a query against factor storage.
type ListQuery struct {
	StockCode   string
	FactorNames []string
	Range       valueobject.DateRange
	DataVersion string
	Limit       int
}
