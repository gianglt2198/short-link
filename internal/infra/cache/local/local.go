package local

import (
	"context"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type entry struct {
	value     []byte
	expiresAt time.Time
}

type LocalCache struct {
	lru *lru.Cache[string, entry]
}

func New(size int) (*LocalCache, error) {
	if size <= 0 {
		size = 1000
	}
	c, err := lru.New[string, entry](size)
	if err != nil {
		return nil, err
	}
	return &LocalCache{lru: c}, nil
}

func (c *LocalCache) Get(_ context.Context, key string) ([]byte, error) {
	e, ok := c.lru.Get(key)
	if !ok {
		return nil, nil
	}
	if time.Now().After(e.expiresAt) {
		c.lru.Remove(key)
		return nil, nil
	}
	return e.value, nil
}

func (c *LocalCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	c.lru.Add(key, entry{value: value, expiresAt: time.Now().Add(ttl)})
	return nil
}

func (c *LocalCache) Close() error { return nil }
