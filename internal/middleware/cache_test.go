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

// mockCacheStore is an in-memory implementation of middleware.CacheStore for testing.
type mockCacheStore struct {
	data   map[string][]byte
	getErr error
	setErr error
}

func newMockCacheStore() *mockCacheStore {
	return &mockCacheStore{data: make(map[string][]byte)}
}

func (m *mockCacheStore) Get(_ context.Context, key string) ([]byte, bool, error) {
	if m.getErr != nil {
		return nil, false, m.getErr
	}
	v, ok := m.data[key]
	return v, ok, nil
}

func (m *mockCacheStore) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}

func TestCache_NonGetRequest_SkipsCache(t *testing.T) {
	store := newMockCacheStore()
	log := zerolog.Nop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusCreated)
	})

	handler := mw.Cache(store, mw.CacheConfig{TTL: time.Minute}, log)(next)

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		nextCalled = false
		req := httptest.NewRequest(method, "/api/resource", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.True(t, nextCalled, "next not called for method %s", method)
		assert.Empty(t, store.data)
	}
}

func TestCache_GetRequest_CacheMissThenHit(t *testing.T) {
	store := newMockCacheStore()
	log := zerolog.Nop()

	callCount := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"value"}`))
	})

	handler := mw.Cache(store, mw.CacheConfig{TTL: time.Minute}, log)(next)

	// First call: cache miss, upstream is hit
	req1 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, "MISS", rr1.Header().Get("X-Cache"))
	assert.Equal(t, 1, callCount)

	// Second call: cache hit, upstream is NOT hit
	req2 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, "HIT", rr2.Header().Get("X-Cache"))
	assert.Equal(t, 1, callCount, "upstream should not be called again on cache HIT")
	assert.Equal(t, `{"data":"value"}`, rr2.Body.String())
}

func TestCache_CacheKey_IncludesQueryString(t *testing.T) {
	store := newMockCacheStore()
	log := zerolog.Nop()

	callCount := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	handler := mw.Cache(store, mw.CacheConfig{TTL: time.Minute}, log)(next)

	// Two requests with different query strings should produce separate cache entries
	req1 := httptest.NewRequest(http.MethodGet, "/api/items?page=1", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/api/items?page=2", nil)

	handler.ServeHTTP(httptest.NewRecorder(), req1)
	handler.ServeHTTP(httptest.NewRecorder(), req2)

	assert.Equal(t, 2, callCount, "cache keys should differ by query string")
	assert.Len(t, store.data, 2)
}

func TestCache_NoCacheHeader_BypassesCache(t *testing.T) {
	store := newMockCacheStore()
	// Pre-populate the cache
	store.data["rc:/cached"] = []byte(`{"cached":true}`)
	log := zerolog.Nop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Cache(store, mw.CacheConfig{TTL: time.Minute}, log)(next)

	req := httptest.NewRequest(http.MethodGet, "/cached", nil)
	req.Header.Set("Cache-Control", "no-cache")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, nextCalled, "upstream should be called when Cache-Control: no-cache")
	assert.Equal(t, "MISS", rr.Header().Get("X-Cache"))
}

func TestCache_StoreReadError_FailsOpen(t *testing.T) {
	store := newMockCacheStore()
	store.getErr = errors.New("redis: connection refused")
	log := zerolog.Nop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Cache(store, mw.CacheConfig{TTL: time.Minute}, log)(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Fail-open: request still reaches upstream
	assert.True(t, nextCalled)
	assert.Equal(t, "MISS", rr.Header().Get("X-Cache"))
}

func TestCache_NonSuccessResponse_NotCached(t *testing.T) {
	store := newMockCacheStore()
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"not_found"}`))
	})

	handler := mw.Cache(store, mw.CacheConfig{TTL: time.Minute}, log)(next)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Empty(t, store.data, "404 responses should not be cached")
}
