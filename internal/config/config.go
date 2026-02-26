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

	LogLevel string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		CoreServiceURL: getEnv("CORE_SERVICE_URL", "https://api.huuhoang.id.vn"),
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
