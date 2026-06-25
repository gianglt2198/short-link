package cache

import (
	"context"
	"time"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/infra/cache/local"
	rediscache "github.com/gianglt1/short-link/internal/infra/cache/redis"
	"go.uber.org/fx"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	Close() error
}

type noopCache struct{}

func (noopCache) Get(_ context.Context, _ string) ([]byte, error)              { return nil, nil }
func (noopCache) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error { return nil }
func (noopCache) Close() error                                                  { return nil }

func NewCache(cfg *config.Config, lc fx.Lifecycle) (Cache, error) {
	if !cfg.Cache.Enabled {
		return noopCache{}, nil
	}

	switch cfg.Cache.Type {
	case "redis":
		c, err := rediscache.New(cfg.Cache.Redis)
		if err != nil {
			return nil, err
		}
		lc.Append(fx.Hook{OnStop: func(_ context.Context) error { return c.Close() }})
		return c, nil
	default:
		return local.New(cfg.Cache.Local.Size)
	}
}

var Module = fx.Provide(NewCache)
