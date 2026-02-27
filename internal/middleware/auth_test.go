package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "github.com/FPT-OJT/gateway/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func signToken(t *testing.T, key *rsa.PrivateKey, subject string, expiry time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": subject,
		"exp": expiry.Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func TestJWTAuth_NoAuthHeader_PassesThrough(t *testing.T) {
	key := generateRSAKey(t)
	log := zerolog.Nop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.JWTAuth(&key.PublicKey, log)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestJWTAuth_ValidToken_SetsXUserIdHeader(t *testing.T) {
	key := generateRSAKey(t)
	log := zerolog.Nop()

	var capturedUserID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.JWTAuth(&key.PublicKey, log)(next)
	tokenStr := signToken(t, key, "user-42", time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "user-42", capturedUserID)
}

func TestJWTAuth_ValidToken_InjectsSubjectIntoContext(t *testing.T) {
	key := generateRSAKey(t)
	log := zerolog.Nop()

	var contextUserID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v, ok := r.Context().Value(mw.UserContextKey{}).(string); ok {
			contextUserID = v
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.JWTAuth(&key.PublicKey, log)(next)
	tokenStr := signToken(t, key, "ctx-user", time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, "ctx-user", contextUserID)
}

func TestJWTAuth_ExpiredToken_Returns401(t *testing.T) {
	key := generateRSAKey(t)
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.JWTAuth(&key.PublicKey, log)(next)
	tokenStr := signToken(t, key, "user-expired", time.Now().Add(-time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "unauthorized")
}

func TestJWTAuth_MalformedHeader_Returns401(t *testing.T) {
	key := generateRSAKey(t)
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.JWTAuth(&key.PublicKey, log)(next)

	for _, badHeader := range []string{"Bearer", "Token abc123", "justtoken"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", badHeader)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code, "header: %s", badHeader)
	}
}

func TestJWTAuth_WrongSigningKey_Returns401(t *testing.T) {
	rightKey := generateRSAKey(t)
	wrongKey := generateRSAKey(t)
	log := zerolog.Nop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Token signed with wrongKey but verified against rightKey's public key
	handler := mw.JWTAuth(&rightKey.PublicKey, log)(next)
	tokenStr := signToken(t, wrongKey, "user-99", time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestJWTAuth_BearerCaseInsensitive_Succeeds(t *testing.T) {
	key := generateRSAKey(t)
	log := zerolog.Nop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.JWTAuth(&key.PublicKey, log)(next)
	tokenStr := signToken(t, key, "user-1", time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}
