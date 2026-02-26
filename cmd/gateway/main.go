package main

import (
	"crypto/rsa"
	"log"
	"os"

	"github.com/FPT-OJT/gateway/internal/cache"
	"github.com/FPT-OJT/gateway/internal/config"
	"github.com/FPT-OJT/gateway/internal/server"
	"github.com/FPT-OJT/gateway/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
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
		Str("redis_url", cfg.RedisURL).
		Msg("configuration loaded")

	rdb, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	defer rdb.Close()

	log.Info().Str("redis_url", cfg.RedisURL).Msg("redis connected")

	pubKey, err := loadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse public key from public.pem")
	}
	log.Info().Msg("public key loaded for JWT verification")

	store := cache.NewRedisStore(rdb)
	router := server.NewRouter(cfg, store, store, pubKey, log)

	srv := server.New(":"+cfg.Port, router, log)
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("server exited with error")
	}
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPublicKeyFromPEM(bytes)
}
