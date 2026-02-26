package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/FPT-OJT/gateway/pkg/errors"
	"github.com/rs/zerolog"
)

// incoming:  GET /api/auth/login
// forwarded: GET auth-endpoint/login
func NewAuthProxy(target *url.URL, log zerolog.Logger) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(target)

	defaultDirector := rp.Director
	rp.Director = func(req *http.Request) {
		defaultDirector(req)

		req.URL.Path = stripPrefix("/api/auth", req.URL.Path)
		req.URL.RawPath = stripPrefix("/api/auth", req.URL.RawPath)

		req.Host = target.Host

		forwardIP(req)

		log.Debug().
			Str("upstream", req.URL.String()).
			Str("method", req.Method).
			Msg("proxying request")
	}

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().
			Err(err).
			Str("upstream", target.String()).
			Str("path", r.URL.Path).
			Msg("upstream request failed")

		errors.WriteJSON(w, http.StatusBadGateway, errors.ErrorResponse{
			Code:    errors.ErrBadGateway.Code,
			Message: errors.ErrBadGateway.Message,
			Detail:  fmt.Sprintf("upstream error: %v", err),
		})
	}

	return rp
}
