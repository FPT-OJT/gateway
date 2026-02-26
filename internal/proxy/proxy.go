package proxy

import (
	"net/http"
	"strings"
)

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

func forwardIP(req *http.Request) {
	if clientIP := realIP(req); clientIP != "" {
		req.Header.Set("X-Forwarded-For", clientIP)
		req.Header.Set("X-Real-IP", clientIP)
	}
}
