package middleware

import (
	"context"
	"time"
)

type RateLimiterStore interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

type CacheStore interface {
	Get(ctx context.Context, key string) (data []byte, found bool, err error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}
