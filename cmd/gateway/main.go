package main

import (
	"log"

	"github.com/FPT-OJT/gateway/internal/config"
	"github.com/FPT-OJT/gateway/internal/server"
	"github.com/FPT-OJT/gateway/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log := logger.New(cfg.LogLevel)
	log.Info().
		Str("port", cfg.Port).
		Str("core_service_url", cfg.CoreServiceURL).
		Msg("configuration loaded")

	router := server.NewRouter(cfg, log)

	srv := server.New(":"+cfg.Port, router, log)
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("server exited with error")
	}
}
