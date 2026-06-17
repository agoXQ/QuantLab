package cache

import (
	"context"
	"sync"
	"time"

	domainCache "github.com/agoXQ/QuantLab/app/formula/domain/cache"
)

type memoryEntry struct {
	data    []byte
	expires time.Time
}

type memoryCache struct {
	mu   sync.RWMutex
	data map[string]memoryEntry
	ttl  time.Duration
}

// NewMemory creates a new in-memory cache (for testing or standalone use).
func NewMemory(ttl time.Duration) domainCache.Cache {
	if ttl == 0 {
		ttl = defaultTTL
	}
	return &memoryCache{
		data: make(map[string]memoryEntry),
		ttl:  ttl,
	}
}

func (c *memoryCache) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok {
		return nil, nil
	}
	if !entry.expires.IsZero() && time.Now().After(entry.expires) {
		delete(c.data, key)
		return nil, nil
	}
	return entry.data, nil
}

func (c *memoryCache) Set(_ context.Context, key string, data []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.ttl
	}
	c.data[key] = memoryEntry{
		data:    data,
		expires: time.Now().Add(ttl),
	}
	return nil
}

func (c *memoryCache) Del(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}
