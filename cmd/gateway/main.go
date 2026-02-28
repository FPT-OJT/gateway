package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/FPT-OJT/gateway/internal/cache"
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
		Str("redis_url", cfg.RedisURL).
		Msg("configuration loaded")

	rdb, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	defer rdb.Close()

	log.Info().Str("redis_url", cfg.RedisURL).Msg("redis connected")

	pubKey, err := loadPublicKey(cfg.PublicKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse public key")
	}
	log.Info().Msg("public key loaded for JWT verification")

	store := cache.NewRedisStore(rdb)
	router := server.NewRouter(cfg, store, store, pubKey, log)

	srv := server.New(":"+cfg.Port, router, log)
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("server exited with error")
	}
}

func loadPublicKey(key string) (*rsa.PublicKey, error) {
	// Decode base64 string
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 public key: %w", err)
	}

	// Try to parse as X.509 PKIX public key first
	pub, err := x509.ParsePKIXPublicKey(decoded)
	if err == nil {
		// Assert it's an RSA public key
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key is not RSA key")
		}
		return rsaPub, nil
	}

	// If PKIX fails, try PKCS1 format
	rsaPub, err := x509.ParsePKCS1PublicKey(decoded)
	if err == nil {
		return rsaPub, nil
	}

	return nil, fmt.Errorf("failed to parse public key in any known format (PKIX or PKCS1): %w", err)
}
