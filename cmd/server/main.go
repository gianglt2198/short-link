package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/handlers"
	"github.com/gianglt1/short-link/internal/helpers"
	"github.com/gianglt1/short-link/internal/infra/database"
	"github.com/gianglt1/short-link/internal/infra/logging"
	"github.com/gianglt1/short-link/internal/repositories"
	"github.com/gianglt1/short-link/internal/services"
)

type serverParams struct {
	fx.In

	App       *fiber.App
	Config    *config.Config
	Lifecycle fx.Lifecycle
}

func startServer(p serverParams) {
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go p.App.Listen(":" + p.Config.Server.Port) //nolint:errcheck
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return p.App.Shutdown()
		},
	})
}

func main() {
	fx.New(
		config.Module,
		logging.Module,
		database.Module,     // *pgxpool.Pool (+ migrations, close hook)
		repositories.Module, // repositories.LinkRepository
		helpers.Module,      // domain.CodeGenerator
		services.Module,
		handlers.Module, // fiber.App + handlers + RegisterRoutes
		fx.WithLogger(func(logger *logging.Logger) fxevent.Logger {
			return logger.Fx()
		}),
		fx.Invoke(startServer),
	).Run()
}
