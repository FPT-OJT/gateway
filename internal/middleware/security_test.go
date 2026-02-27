package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/FPT-OJT/gateway/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestSecurity_SetsAllSecurityHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw.Security(next).ServeHTTP(rr, req)

	h := rr.Header()
	assert.Equal(t, "nosniff", h.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", h.Get("X-Frame-Options"))
	assert.Contains(t, h.Get("Strict-Transport-Security"), "max-age=31536000")
	assert.Contains(t, h.Get("Strict-Transport-Security"), "includeSubDomains")
	assert.Equal(t, "0", h.Get("X-XSS-Protection"))
	assert.Equal(t, "no-referrer", h.Get("Referrer-Policy"))
}

func TestSecurity_SetsCORSHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw.Security(next).ServeHTTP(rr, req)

	h := rr.Header()
	assert.Equal(t, "*", h.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, h.Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, h.Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, h.Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Contains(t, h.Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestSecurity_OptionsRequest_Returns204_AndDoesNotCallNext(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/resource", nil)
	rr := httptest.NewRecorder()
	mw.Security(next).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.False(t, nextCalled, "next handler must not be invoked for OPTIONS preflight")
}

func TestSecurity_NonOptionsRequest_CallsNext(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw.Security(next).ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}
