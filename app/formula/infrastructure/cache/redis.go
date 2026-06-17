package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	domainCache "github.com/agoXQ/QuantLab/app/formula/domain/cache"
)

const (
	defaultTTL = 24 * time.Hour
)

type redisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedis creates a new Redis-backed cache.
func NewRedis(client *redis.Client, ttl time.Duration) domainCache.Cache {
	if ttl == 0 {
		ttl = defaultTTL
	}
	return &redisCache{
		client: client,
		ttl:    ttl,
	}
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

func (c *redisCache) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set %s: %w", key, err)
	}
	return nil
}

func (c *redisCache) Del(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del %s: %w", key, err)
	}
	return nil
}
