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

func NewRouter(cfg *config.Config, log zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	initMiddleware(r, log)

	r.Get("/health", handleHealth)
	mountProxy(r, cfg, log)

	return r
}

func initMiddleware(r *chi.Mux, log zerolog.Logger) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.Recovery(log))
	r.Use(mw.Security)
	r.Use(mw.TraceLog(log))
	r.Use(middleware.Compress(5))
}

func mountProxy(r *chi.Mux, cfg *config.Config, log zerolog.Logger) {
	coreTarget, _ := url.Parse(cfg.CoreServiceURL)
	coreProxy := proxy.NewCoreProxy(coreTarget, log)
	authTarget, _ := url.Parse(cfg.AuthServiceURL)
	authProxy := proxy.NewAuthProxy(authTarget, log)
	aiTarget, _ := url.Parse(cfg.AiServiceURL)
	aiProxy := proxy.NewAiProxy(aiTarget, log)

	r.Mount("/api/core", http.StripPrefix("/api/core", coreProxy))
	r.Mount("/api/auth", http.StripPrefix("/api/auth", authProxy))
	r.Mount("/api/ai", http.StripPrefix("/api/ai", aiProxy))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
