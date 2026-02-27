package cache_test

import (
	"testing"

	"github.com/FPT-OJT/gateway/internal/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisStore_ReturnsNonNil(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	store := cache.NewRedisStore(client)
	assert.NotNil(t, store)
}
