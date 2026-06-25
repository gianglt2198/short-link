package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/fx"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/infra/logging"
)

type PoolParams struct {
	fx.In

	Config    *config.Config
	Logger    *logging.Logger
	Lifecycle fx.Lifecycle
}

func NewPool(p PoolParams) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(p.Config.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   p.Logger,
		LogLevel: tracelog.LogLevelDebug,
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	if err := runMigrations(p.Config.Database.URL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}

func runMigrations(databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

var Module = fx.Provide(NewPool)
