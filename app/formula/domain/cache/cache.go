package cache

import (
	"context"
	"time"
)

// Cache defines a generic key-value cache interface.
// Implementations must be safe for concurrent use.
type Cache interface {
	// Get retrieves a value by key. Returns nil, nil if the key does not exist.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with a TTL.
	Set(ctx context.Context, key string, data []byte, ttl time.Duration) error

	// Del removes a key.
	Del(ctx context.Context, key string) error
}
