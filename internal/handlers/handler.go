package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

func RegisterRoutes(app *fiber.App, lh *LinkHandler) {
	app.Post("/encode", lh.Encode)
	app.Post("/decode", lh.Decode)
	app.Get("/:code", lh.Redirect)
}

func NewApp() *fiber.App {
	return fiber.New(fiber.Config{BodyLimit: 4 * 1024})
}

var Module = fx.Options(
	fx.Provide(NewApp),
	fx.Provide(NewLinkHandler),
	fx.Invoke(RegisterRoutes),
)
