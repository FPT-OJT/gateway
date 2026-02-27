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

// TestRecovery_PanicWithNil_DoesNotCrash verifies the recovery middleware
// handles panic(nil) without crashing the server.
// The response code is intentionally not asserted: Go ≥1.21 recovers
// panic(nil) as *runtime.PanicNilError (non-nil → 500), while older
// toolchains recover it as nil (middleware skips → 200). Either is safe.
func TestRecovery_PanicWithNil_DoesNotCrash(t *testing.T) {
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(nil)
	})

	handler := mw.Recovery(log)(next)
	req := httptest.NewRequest(http.MethodGet, "/nil-panic", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(rr, req)
	})

	// Code is either 200 (nil recovered → no action) or 500 (PanicNilError recovered).
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, rr.Code)
}
