package server

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/FPT-OJT/gateway/internal/config"
	mw "github.com/FPT-OJT/gateway/internal/middleware"
	"github.com/FPT-OJT/gateway/internal/proxy"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func NewRouter(cfg *config.Config, rateStore mw.RateLimiterStore, cacheStore mw.CacheStore, log zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	initMiddleware(r, cfg, rateStore, cacheStore, log)

	r.Get("/health", handleHealth)
	mountProxy(r, cfg, log)

	return r
}

func initMiddleware(r *chi.Mux, cfg *config.Config, rateStore mw.RateLimiterStore, cacheStore mw.CacheStore, log zerolog.Logger) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.Recovery(log))
	r.Use(mw.Security)
	r.Use(mw.TraceLog(log))
	r.Use(middleware.Compress(5))

	r.Use(mw.RateLimit(rateStore, mw.RateLimitConfig{
		RPS:   cfg.RateLimitRPS,
		Burst: cfg.RateLimitBurst,
	}, log))

	r.Use(mw.Cache(cacheStore, mw.CacheConfig{
		TTL: cfg.CacheTTL,
	}, log))
}

func mountProxy(r *chi.Mux, cfg *config.Config, log zerolog.Logger) {
	services := []struct {
		prefix string
		rawURL string
	}{
		{"/api/core", cfg.CoreServiceURL},
		{"/api/auth", cfg.AuthServiceURL},
		{"/api/ai", cfg.AiServiceURL},
	}

	for _, svc := range services {
		target, _ := url.Parse(svc.rawURL)
		p := proxy.New(proxy.Config{Prefix: svc.prefix, Target: target}, log)
		r.Mount(svc.prefix, http.StripPrefix(svc.prefix, p))
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
