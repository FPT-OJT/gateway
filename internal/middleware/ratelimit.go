// Package middleware provides reusable HTTP middleware for the API gateway.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/FPT-OJT/gateway/pkg/utils"
	"github.com/rs/zerolog"
)

type RateLimitConfig struct {
	RPS   int
	Burst int
}

// RateLimit returns a middleware that enforces a per-IP sliding-window rate
// limit backed by a RateLimiterStore.
//
// Algorithm: Store INCR + EXPIRE per (IP, second) key.
//   - First request in a window: INCR → 1, then EXPIRE sets TTL to 2 s.
//   - Subsequent requests: INCR increments the counter.
//   - If counter > RPS+Burst → 429 Too Many Requests.
//
// The 2-second TTL gives a small grace window across second boundaries while
// keeping memory use bounded.
func RateLimit(store RateLimiterStore, cfg RateLimitConfig, log zerolog.Logger) func(http.Handler) http.Handler {
	limit := cfg.RPS + cfg.Burst

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := utils.ClientIp(r)
			key := fmt.Sprintf("rl:%s:%d", ip, time.Now().UTC().Unix())

			ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
			defer cancel()

			count, err := store.Incr(ctx, key)
			if err != nil {
				log.Warn().Err(err).Str("ip", ip).Msg("rate limit: store error, failing open")
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				_ = store.Expire(ctx, key, 2*time.Second)
			}

			if int(count) > limit {
				log.Warn().
					Str("ip", ip).
					Int64("count", count).
					Int("limit", limit).
					Msg("rate limit exceeded")

				w.Header().Set("Retry-After", "1")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, `{"code":"rate_limited","message":"Too many requests"}`, http.StatusTooManyRequests)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)))
			next.ServeHTTP(w, r)
		})
	}
}
