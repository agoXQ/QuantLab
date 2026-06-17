// Package dataversion defines the DataVersion aggregate.
//
// Every fact stored by the Market Data service is bound to a DataVersion so
// downstream services (e.g. backtest) can reproduce historical results.
package dataversion

import (
	"context"
	"time"
)

// DataVersion describes a snapshot of market data.
type DataVersion struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Repository persists DataVersion records.
type Repository interface {
	Get(ctx context.Context, version string) (*DataVersion, error)
	List(ctx context.Context, limit int) ([]*DataVersion, error)
	Latest(ctx context.Context) (*DataVersion, error)
	Create(ctx context.Context, dv *DataVersion) error
}
