package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	CoreServiceURL string
	AuthServiceURL string
	AiServiceURL   string

	LogLevel string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		CoreServiceURL: getEnv("CORE_SERVICE_URL", "http://localhost:8090"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8091"),
		AiServiceURL:   getEnv("AI_SERVICE_URL", "http://localhost:8092"),
		LogLevel:       strings.ToLower(getEnv("LOG_LEVEL", "info")),
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

	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
