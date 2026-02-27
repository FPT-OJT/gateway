package cache_test

import (
	"testing"

	"github.com/FPT-OJT/gateway/internal/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// TestNewRedisStore_ReturnsNonNil verifies the constructor returns a non-nil store
// without requiring a live Redis connection.
func TestNewRedisStore_ReturnsNonNil(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	store := cache.NewRedisStore(client)
	assert.NotNil(t, store)
}

// TestNewRedisStore_DifferentClients verifies each call returns a distinct store.
func TestNewRedisStore_DifferentClients(t *testing.T) {
	client1 := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	client2 := redis.NewClient(&redis.Options{Addr: "localhost:6380"})
	defer client1.Close()
	defer client2.Close()

	store1 := cache.NewRedisStore(client1)
	store2 := cache.NewRedisStore(client2)

	assert.NotNil(t, store1)
	assert.NotNil(t, store2)
	assert.NotEqual(t, store1, store2)
}
