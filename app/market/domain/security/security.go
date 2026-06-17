// Package security defines the Security aggregate root for the Market Data service.
package security

import (
	"context"
	"strings"
	"time"

	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// Security is the aggregate root describing a tradable instrument.
type Security struct {
	ID            int64                    `json:"id"`
	StockCode     string                   `json:"stock_code"`
	StockName     string                   `json:"stock_name"`
	Market        valueobject.Market       `json:"market"`
	Exchange      string                   `json:"exchange"`
	AssetType     valueobject.AssetType    `json:"asset_type"`
	Industry      string                   `json:"industry"`
	ListingDate   time.Time                `json:"listing_date"`
	DelistingDate time.Time                `json:"delisting_date,omitempty"`
	Status        valueobject.SecurityStatus `json:"status"`
}

// Normalize trims whitespace and uppercases identifiers in a stable way.
func (s *Security) Normalize() {
	s.StockCode = strings.ToUpper(strings.TrimSpace(s.StockCode))
	s.Exchange = strings.ToUpper(strings.TrimSpace(s.Exchange))
	s.StockName = strings.TrimSpace(s.StockName)
	s.Industry = strings.TrimSpace(s.Industry)
}

// IsActive reports whether the security is tradable at the given moment.
func (s *Security) IsActive(at time.Time) bool {
	if s.Status == valueobject.StatusDelisted {
		return false
	}
	if !s.DelistingDate.IsZero() && !at.Before(s.DelistingDate) {
		return false
	}
	if !s.ListingDate.IsZero() && at.Before(s.ListingDate) {
		return false
	}
	return true
}

// Repository persists Security aggregates.
//
// Implementations must be safe for concurrent use and return ErrSecurityNotFound
// when a record cannot be located.
type Repository interface {
	GetByCode(ctx context.Context, stockCode string) (*Security, error)
	List(ctx context.Context, query ListQuery) ([]*Security, string, error)
	Upsert(ctx context.Context, sec *Security) error
	BulkUpsert(ctx context.Context, list []*Security) error
}

// ListQuery contains optional filters for listing securities.
type ListQuery struct {
	Market    valueobject.Market
	Exchange  string
	AssetType valueobject.AssetType
	Cursor    string
	Limit     int
}
