package handlers

import (
	"github.com/gianglt1/short-link/internal/infra/logging"
	"github.com/gianglt1/short-link/internal/middlewares"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

type Handler struct {
	link *LinkHandler
}

func NewHandler(link *LinkHandler) *Handler {
	return &Handler{link: link}
}

func RegisterRoutes(app *fiber.App, h *Handler) {
	app.Post("/encode", h.link.Encode)
	app.Post("/decode", h.link.Decode)
	app.Get("/:code", h.link.Redirect)
}

func NewApp(log *logging.Logger) *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 4 * 1024})
	app.Use(middlewares.RequestIDMiddleware)
	app.Use(middlewares.LoggerMiddleware(log))
	return app
}

var Module = fx.Options(
	fx.Provide(NewApp),
	fx.Provide(NewLinkHandler),
	fx.Provide(NewHandler),
	fx.Invoke(RegisterRoutes),
)
