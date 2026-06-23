package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/gianglt1/short-link/internal/config"
)

type PoolParams struct {
	fx.In

	Config    *config.Config
	Lifecycle fx.Lifecycle
}

func NewPool(p PoolParams) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), p.Config.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
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
