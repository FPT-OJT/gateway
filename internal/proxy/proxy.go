package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/FPT-OJT/gateway/pkg/errors"
	"github.com/FPT-OJT/gateway/pkg/utils"

	"github.com/rs/zerolog"
)

type Config struct {
	Prefix string
	Target *url.URL
}

func New(cfg Config, log zerolog.Logger) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(cfg.Target)

	defaultDirector := rp.Director
	rp.Director = func(req *http.Request) {
		defaultDirector(req)
		req.URL.Path = stripPrefix(cfg.Prefix, req.URL.Path)
		req.URL.RawPath = stripPrefix(cfg.Prefix, req.URL.RawPath)
		req.Host = cfg.Target.Host
		forwardIP(req)

		log.Debug().
			Str("upstream", req.URL.String()).
			Str("method", req.Method).
			Msg("proxying request")
	}

	rp.ErrorHandler = makeErrorHandler(cfg.Target, log)
	return rp
}

func makeErrorHandler(target *url.URL, log zerolog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
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

func forwardIP(req *http.Request) {
	if clientIP := utils.ClientIp(req); clientIP != "" {
		req.Header.Set("X-Forwarded-For", clientIP)
		req.Header.Set("X-Real-IP", clientIP)
	}
}
