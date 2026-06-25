package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gianglt1/short-link/internal/config"
	rediscache "github.com/gianglt1/short-link/internal/infra/cache/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCache(t *testing.T) *rediscache.RedisCache {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	c, err := rediscache.New(config.RedisConfig{Addr: addr})
	if err != nil {
		t.Skipf("redis unavailable (%v) — skipping integration tests", err)
	}

	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestRedisCache_Miss(t *testing.T) {
	c := newCache(t)
	val, err := c.Get(context.Background(), "redis:missing:key")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestRedisCache_SetAndGet(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()
	key := "redis:test:setget"

	t.Cleanup(func() { c.Get(ctx, key) }) //nolint:errcheck

	require.NoError(t, c.Set(ctx, key, []byte("hello"), time.Minute))

	val, err := c.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), val)
}

func TestRedisCache_Overwrite(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()
	key := "redis:test:overwrite"

	require.NoError(t, c.Set(ctx, key, []byte("first"), time.Minute))
	require.NoError(t, c.Set(ctx, key, []byte("second"), time.Minute))

	val, err := c.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), val)
}

func TestRedisCache_TTLExpiry(t *testing.T) {
	c := newCache(t)
	ctx := context.Background()
	key := "redis:test:ttl"

	require.NoError(t, c.Set(ctx, key, []byte("ephemeral"), 50*time.Millisecond))
	time.Sleep(100 * time.Millisecond)

	val, err := c.Get(ctx, key)
	require.NoError(t, err)
	assert.Nil(t, val, "expired key should return nil")
}

func TestRedisCache_Close(t *testing.T) {
	c := newCache(t)
	assert.NoError(t, c.Close())
}
