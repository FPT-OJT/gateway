// Package middleware provides reusable HTTP middleware for the API gateway.
package middleware

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type CacheConfig struct {
	TTL time.Duration
}

// Cache returns a middleware that caches upstream GET responses using a CacheStore.
//
// Rules:
//   - Only GET requests are cached.
//   - Responses with non-2xx status are never cached.
//   - Requests with "Cache-Control: no-cache" bypass the cache entirely.
//   - Cache key: "rc:{path}?{rawquery}"
//   - X-Cache: HIT  → served from cache.
//   - X-Cache: MISS → fetched from upstream, then stored.
func Cache(store CacheStore, cfg CacheConfig, log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("Cache-Control") == "no-cache" {
				w.Header().Set("X-Cache", "MISS")
				next.ServeHTTP(w, r)
				return
			}

			key := cacheKey(r)

			ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
			defer cancel()

			cached, found, err := store.Get(ctx, key)
			if found {
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(cached)

				log.Debug().Str("key", key).Msg("cache: HIT")
				return
			}

			if err != nil {
				log.Warn().Err(err).Str("key", key).Msg("cache: store read error, failing open")
				w.Header().Set("X-Cache", "MISS")
				next.ServeHTTP(w, r)
				return
			}

			rec := &responseRecorder{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
				status:         http.StatusOK,
			}

			w.Header().Set("X-Cache", "MISS")
			next.ServeHTTP(rec, r)

			if rec.status >= 200 && rec.status < 300 && rec.buf.Len() > 0 {
				storeCtx, storeCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer storeCancel()

				if setErr := store.Set(storeCtx, key, rec.buf.Bytes(), cfg.TTL); setErr != nil {
					log.Warn().Err(setErr).Str("key", key).Msg("cache: store write error")
				} else {
					log.Debug().Str("key", key).Dur("ttl", cfg.TTL).Msg("cache: stored")
				}
			}
		})
	}
}

// cacheKey builds a stable cache key from the request path and query string.
func cacheKey(r *http.Request) string {
	var sb strings.Builder
	sb.WriteString("rc:")
	sb.WriteString(r.URL.Path)
	if q := r.URL.RawQuery; q != "" {
		sb.WriteByte('?')
		sb.WriteString(q)
	}
	return sb.String()
}

// responseRecorder wraps http.ResponseWriter so we can capture the body and
// status code without preventing the response from reaching the original writer.
type responseRecorder struct {
	http.ResponseWriter
	buf    *bytes.Buffer
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.buf.Write(b) // capture
	return r.ResponseWriter.Write(b)
}
