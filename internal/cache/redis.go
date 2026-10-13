package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(ctx context.Context, url string) (*Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	
	opt.PoolSize = 20
	opt.MinIdleConns = 5
	opt.DialTimeout = 3 * time.Second
	opt.ReadTimeout = 2 * time.Second
	opt.WriteTimeout = 2 * time.Second

	rdb := redis.NewClient(opt)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Raw() *redis.Client { return c.rdb }

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error { return c.rdb.Close() }

func (c *Client) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil 
	}
	if err != nil {
		return false, fmt.Errorf("redis get: %w", err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return false, fmt.Errorf("unmarshal cached: %w", err)
	}
	return true, nil
}

func (c *Client) SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal for cache: %w", err)
	}
	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

func (c *Client) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("redis scan: %w", err)
		}
		if len(keys) > 0 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis del batch: %w", err)
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func LogSafe(op string, err error, args ...any) {
	if err == nil {
		return
	}
	attrs := append([]any{"op", op, "err", err.Error()}, args...)
	slog.Warn("cache_degraded", attrs...)
}