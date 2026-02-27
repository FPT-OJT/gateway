package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/FPT-OJT/gateway/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestRecovery_NoPanic_PassesThrough(t *testing.T) {
	log := zerolog.Nop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Recovery(log)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRecovery_PanicWithString_Returns500(t *testing.T) {
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	handler := mw.Recovery(log)(next)
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(rr, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal_error")
}

func TestRecovery_PanicWithError_Returns500(t *testing.T) {
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(assert.AnError)
	})

	handler := mw.Recovery(log)(next)
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(rr, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestRecovery_PanicWithNil_Returns500(t *testing.T) {
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	handler := mw.Recovery(log)(next)
	req := httptest.NewRequest(http.MethodGet, "/nil-panic", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(rr, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
