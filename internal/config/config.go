package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	CoreServiceURL string
	AuthServiceURL string
	AiServiceURL   string

	LogLevel string

	// Redis connection URL (e.g. redis://:password@host:6379/0).
	RedisURL string

	RateLimitRPS   int
	RateLimitBurst int

	CacheTTL time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	rps, err := strconv.Atoi(getEnv("RATE_LIMIT_RPS", "100"))
	if err != nil {
		return nil, fmt.Errorf("config: RATE_LIMIT_RPS must be an integer: %w", err)
	}

	burst, err := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))
	if err != nil {
		return nil, fmt.Errorf("config: RATE_LIMIT_BURST must be an integer: %w", err)
	}

	ttlSec, err := strconv.Atoi(getEnv("CACHE_TTL", "60"))
	if err != nil {
		return nil, fmt.Errorf("config: CACHE_TTL must be an integer (seconds): %w", err)
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		CoreServiceURL: getEnv("CORE_SERVICE_URL", "http://localhost:8090"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8091"),
		AiServiceURL:   getEnv("AI_SERVICE_URL", "http://localhost:8092"),
		LogLevel:       strings.ToLower(getEnv("LOG_LEVEL", "info")),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/0"),
		RateLimitRPS:   rps,
		RateLimitBurst: burst,
		CacheTTL:       time.Duration(ttlSec) * time.Second,
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.CoreServiceURL == "" {
		return fmt.Errorf("CORE_SERVICE_URL must not be empty")
	}
	if c.AuthServiceURL == "" {
		return fmt.Errorf("AUTH_SERVICE_URL must not be empty")
	}
	if c.AiServiceURL == "" {
		return fmt.Errorf("AI_SERVICE_URL must not be empty")
	}
	if c.Port == "" {
		return fmt.Errorf("PORT must not be empty")
	}
	if c.RateLimitRPS <= 0 {
		return fmt.Errorf("RATE_LIMIT_RPS must be greater than 0")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
