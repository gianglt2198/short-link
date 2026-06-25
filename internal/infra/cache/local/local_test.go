package local_test

import (
	"context"
	"testing"
	"time"

	"github.com/gianglt1/short-link/internal/infra/cache/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCache(t *testing.T, size int) *local.LocalCache {
	t.Helper()
	c, err := local.New(size)
	require.NoError(t, err)
	return c
}

func TestLocalCache_Miss(t *testing.T) {
	c := newCache(t, 10)
	val, err := c.Get(context.Background(), "missing")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestLocalCache_SetAndGet(t *testing.T) {
	c := newCache(t, 10)
	err := c.Set(context.Background(), "key1", []byte("value1"), time.Minute)
	require.NoError(t, err)

	val, err := c.Get(context.Background(), "key1")
	require.NoError(t, err)
	assert.Equal(t, []byte("value1"), val)
}

func TestLocalCache_Overwrite(t *testing.T) {
	c := newCache(t, 10)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "k", []byte("first"), time.Minute))
	require.NoError(t, c.Set(ctx, "k", []byte("second"), time.Minute))

	val, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), val)
}

func TestLocalCache_TTLExpiry(t *testing.T) {
	c := newCache(t, 10)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "expiring", []byte("data"), time.Millisecond))
	time.Sleep(5 * time.Millisecond)

	val, err := c.Get(ctx, "expiring")
	require.NoError(t, err)
	assert.Nil(t, val, "expired entry should return nil")
}

func TestLocalCache_LRUEviction(t *testing.T) {
	c := newCache(t, 2)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "a", []byte("A"), time.Minute))
	require.NoError(t, c.Set(ctx, "b", []byte("B"), time.Minute))
	// access "a" to make "b" the LRU candidate
	_, _ = c.Get(ctx, "a")
	// adding "c" should evict "b"
	require.NoError(t, c.Set(ctx, "c", []byte("C"), time.Minute))

	val, err := c.Get(ctx, "b")
	require.NoError(t, err)
	assert.Nil(t, val, "evicted entry should return nil")

	val, err = c.Get(ctx, "a")
	require.NoError(t, err)
	assert.Equal(t, []byte("A"), val)
}

func TestLocalCache_Close(t *testing.T) {
	c := newCache(t, 10)
	assert.NoError(t, c.Close())
}

func TestLocalCache_DefaultSize(t *testing.T) {
	// size <= 0 should not error
	c, err := local.New(0)
	require.NoError(t, err)
	require.NoError(t, c.Set(context.Background(), "k", []byte("v"), time.Minute))
	val, err := c.Get(context.Background(), "k")
	require.NoError(t, err)
	assert.Equal(t, []byte("v"), val)
}
