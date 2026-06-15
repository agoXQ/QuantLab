// Package redis provides shared Redis utilities for all QuantLab services.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Config holds Redis connection configuration.
type Config struct {
	Addr     string
	Password string
	DB       int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(addr string) Config {
	return Config{
		Addr: addr,
		DB:   0,
	}
}

// NewClient creates a new Redis client.
func NewClient(ctx context.Context, cfg Config) (*goredis.Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Println("[redis] connected")
	return rdb, nil
}

// CloseClient gracefully closes the Redis client.
func CloseClient(rdb *goredis.Client) {
	if rdb != nil {
		rdb.Close()
		log.Println("[redis] closed")
	}
}

// Cache provides typed get/set/invalidate operations.
type Cache struct {
	client *goredis.Client
}

// NewCache creates a new Cache wrapper.
func NewCache(client *goredis.Client) *Cache {
	return &Cache{client: client}
}

// GetJSON retrieves and unmarshals a cached value into dest.
// Returns false if the key does not exist.
func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cache get %s: %w", key, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("cache unmarshal %s: %w", key, err)
	}
	return true, nil
}

// SetJSON marshals and caches a value with the given TTL.
func (c *Cache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal %s: %w", key, err)
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes one or more keys.
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists.
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Incr atomically increments a counter.
func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}
