// Package cache defines a generic key-value cache abstraction shared across
// the Market Data application layer.
package cache

import (
	"context"
	"time"
)

// Cache provides typed JSON-friendly key-value access. Implementations must be
// safe for concurrent use.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}
