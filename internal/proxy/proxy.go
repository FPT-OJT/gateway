package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/FPT-OJT/gateway/pkg/errors"
	"github.com/rs/zerolog"
)

// incoming:  GET /api/core/users/1
// forwarded: GET https://api.huuhoang.id.vn/users/1
func NewCoreProxy(target *url.URL, log zerolog.Logger) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(target)

	defaultDirector := rp.Director
	rp.Director = func(req *http.Request) {
		defaultDirector(req)

		req.URL.Path = stripPrefix("/api/core", req.URL.Path)
		req.URL.RawPath = stripPrefix("/api/core", req.URL.RawPath)

		req.Host = target.Host

		if clientIP := realIP(req); clientIP != "" {
			req.Header.Set("X-Forwarded-For", clientIP)
			req.Header.Set("X-Real-IP", clientIP)
		}

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

func stripPrefix(prefix, s string) string {
	if s == "" {
		return s
	}
	trimmed := strings.TrimPrefix(s, prefix)
	if trimmed == "" {
		return "/"
	}
	return trimmed
}

func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.SplitN(fwd, ",", 2)[0]
	}
	return r.RemoteAddr
}
