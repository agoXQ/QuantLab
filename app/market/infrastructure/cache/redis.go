// Package cache provides a Redis-backed implementation of domain.Cache.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	domainCache "github.com/agoXQ/QuantLab/app/market/domain/cache"
)

const defaultTTL = 1 * time.Hour

type redisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedis returns a domain.Cache backed by go-redis.
func NewRedis(client *redis.Client, ttl time.Duration) domainCache.Cache {
	if ttl == 0 {
		ttl = defaultTTL
	}
	return &redisCache{client: client, ttl: ttl}
}

func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get %s: %w", key, err)
	}
	return data, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set %s: %w", key, err)
	}
	return nil
}

func (c *redisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}
