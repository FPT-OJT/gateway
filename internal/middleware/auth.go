// Package middleware provides reusable HTTP middleware for the API gateway.
package middleware

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

// UserContextKey is used to store the authenticated user ID in the request context.
type UserContextKey struct{}

// JWTAuth returns a middleware that verifies RS256 JWTs in the Authorization
// header using the provided RSA public key.
//
// Behavior:
//   - If NO Authorization header is present, it allows the request through
//     unauthenticated (upstream services handle public routes).
//   - If a Bearer token IS present, it must be valid and unexpired.
//   - If invalid, returns 401 Unauthorized immediately.
//   - If valid, extracts the "sub" claim, injects it into the request context,
//     and adds the "X-User-Id" HTTP header for upstream services.
func JWTAuth(pubKey *rsa.PublicKey, log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for public routes
			if strings.Contains(r.URL.Path, "/public") {
				log.Debug().Str("path", r.URL.Path).Msg("auth: public route, skipping authentication")
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Debug().Msg("auth: no authorization header, proceeding unauthenticated")
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				log.Warn().Str("header", authHeader).Msg("auth: malformed authorization header")
				sendUnauthorized(w, "Malformed Authorization header. Expected 'Bearer <token>'")
				return
			}
			tokenStr := parts[1]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return pubKey, nil
			})

			if err != nil || !token.Valid {
				log.Warn().Err(err).Msg("auth: invalid or expired token")
				sendUnauthorized(w, "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Error().Msg("auth: failed to extract claims from valid token")
				sendUnauthorized(w, "Invalid token claims")
				return
			}

			sub, err := claims.GetSubject()
			if err != nil || sub == "" {
				log.Warn().Msg("auth: token missing 'sub' claim")
				sendUnauthorized(w, "Token is missing subject (sub) claim")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey{}, sub)
			r = r.WithContext(ctx)

			r.Header.Set("X-User-Id", sub)

			log.Debug().Str("sub", sub).Msg("auth: successfully authenticated")

			next.ServeHTTP(w, r)
		})
	}
}

// sendUnauthorized writes a 401 JSON response.
func sendUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"code":"unauthorized","message":%q}`, message)
}
