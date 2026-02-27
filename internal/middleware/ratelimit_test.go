package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "github.com/FPT-OJT/gateway/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// mockRateLimiterStore is an in-memory implementation of middleware.RateLimiterStore.
type mockRateLimiterStore struct {
	counters map[string]int64
	incrErr  error
}

func newMockRateLimiterStore() *mockRateLimiterStore {
	return &mockRateLimiterStore{counters: make(map[string]int64)}
}

func (m *mockRateLimiterStore) Incr(_ context.Context, key string) (int64, error) {
	if m.incrErr != nil {
		return 0, m.incrErr
	}
	m.counters[key]++
	return m.counters[key], nil
}

func (m *mockRateLimiterStore) Expire(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func TestRateLimit_UnderLimit_Passes(t *testing.T) {
	store := newMockRateLimiterStore()
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.RateLimit(store, mw.RateLimitConfig{RPS: 10, Burst: 5}, log)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "15", rr.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "14", rr.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimit_ExceedsLimit_Returns429(t *testing.T) {
	store := newMockRateLimiterStore()
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// RPS=2, Burst=1 → limit = 3; 4th request should be rate-limited
	handler := mw.RateLimit(store, mw.RateLimitConfig{RPS: 2, Burst: 1}, log)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"

	// Exhaust the limit
	for i := 0; i < 3; i++ {
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	// 4th request should be rejected
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Equal(t, "1", rr.Header().Get("Retry-After"))
	assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Remaining"))
	assert.Contains(t, rr.Body.String(), "rate_limited")
}

func TestRateLimit_StoreError_FailsOpen(t *testing.T) {
	store := newMockRateLimiterStore()
	store.incrErr = errors.New("redis: dial tcp: connection refused")
	log := zerolog.Nop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.RateLimit(store, mw.RateLimitConfig{RPS: 10, Burst: 5}, log)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Fail-open: request should still reach upstream
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimit_DifferentIPs_IndependentCounters(t *testing.T) {
	store := newMockRateLimiterStore()
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Limit of 1 (RPS=1, Burst=0)
	handler := mw.RateLimit(store, mw.RateLimitConfig{RPS: 1, Burst: 0}, log)(next)

	// IP A exhausts its quota
	reqA := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA.RemoteAddr = "192.168.1.1:1111"
	handler.ServeHTTP(httptest.NewRecorder(), reqA)
	rrA2 := httptest.NewRecorder()
	handler.ServeHTTP(rrA2, reqA)
	assert.Equal(t, http.StatusTooManyRequests, rrA2.Code)

	// IP B should still be allowed
	reqB := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB.RemoteAddr = "192.168.1.2:2222"
	rrB := httptest.NewRecorder()
	handler.ServeHTTP(rrB, reqB)
	assert.Equal(t, http.StatusOK, rrB.Code)
}
