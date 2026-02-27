package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/FPT-OJT/gateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearEnv removes all config-related environment variables to prevent bleed-over between tests.
func clearEnv(t *testing.T) {
	t.Helper()
	vars := []string{
		"PORT", "CORE_SERVICE_URL", "AUTH_SERVICE_URL", "AI_SERVICE_URL",
		"REDIS_URL", "RATE_LIMIT_RPS", "RATE_LIMIT_BURST",
		"CACHE_TTL", "LOG_LEVEL", "PUBLIC_KEY_PATH",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	clearEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "http://localhost:8090", cfg.CoreServiceURL)
	assert.Equal(t, "http://localhost:8091", cfg.AuthServiceURL)
	assert.Equal(t, "http://localhost:8092", cfg.AiServiceURL)
	assert.Equal(t, "redis://localhost:6379/0", cfg.RedisURL)
	assert.Equal(t, 100, cfg.RateLimitRPS)
	assert.Equal(t, 20, cfg.RateLimitBurst)
	assert.Equal(t, 60*time.Second, cfg.CacheTTL)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "public.pem", cfg.PublicKeyPath)
}

func TestLoad_OverrideWithEnvVars(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("CORE_SERVICE_URL", "http://core:8080")
	t.Setenv("AUTH_SERVICE_URL", "http://auth:8080")
	t.Setenv("AI_SERVICE_URL", "http://ai:8080")
	t.Setenv("RATE_LIMIT_RPS", "50")
	t.Setenv("RATE_LIMIT_BURST", "10")
	t.Setenv("CACHE_TTL", "30")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("PUBLIC_KEY_PATH", "key.pem")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "http://core:8080", cfg.CoreServiceURL)
	assert.Equal(t, "http://auth:8080", cfg.AuthServiceURL)
	assert.Equal(t, "http://ai:8080", cfg.AiServiceURL)
	assert.Equal(t, 50, cfg.RateLimitRPS)
	assert.Equal(t, 10, cfg.RateLimitBurst)
	assert.Equal(t, 30*time.Second, cfg.CacheTTL)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "key.pem", cfg.PublicKeyPath)
}

func TestLoad_LogLevel_NormalisedToLower(t *testing.T) {
	clearEnv(t)
	t.Setenv("LOG_LEVEL", "WARN")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "warn", cfg.LogLevel)
}

func TestLoad_InvalidRateLimitRPS_ReturnsError(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_RPS", "not-a-number")

	_, err := config.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RATE_LIMIT_RPS")
}

func TestLoad_InvalidRateLimitBurst_ReturnsError(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_BURST", "invalid")

	_, err := config.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RATE_LIMIT_BURST")
}

func TestLoad_InvalidCacheTTL_ReturnsError(t *testing.T) {
	clearEnv(t)
	t.Setenv("CACHE_TTL", "invalid")

	_, err := config.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CACHE_TTL")
}

func TestLoad_ZeroRPS_FailsValidation(t *testing.T) {
	clearEnv(t)
	t.Setenv("RATE_LIMIT_RPS", "0")

	_, err := config.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RATE_LIMIT_RPS")
}
