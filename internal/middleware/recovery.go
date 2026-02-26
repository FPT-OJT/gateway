package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/FPT-OJT/gateway/pkg/errors"
	"github.com/rs/zerolog"
)

func Recovery(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error().
						Str("panic", fmt.Sprintf("%v", rec)).
						Str("stack", string(debug.Stack())).
						Str("path", r.URL.Path).
						Msg("recovered from panic")

					errors.WriteJSON(w, http.StatusInternalServerError, errors.ErrInternal)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
