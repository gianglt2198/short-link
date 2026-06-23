package handlers

import (
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

func NewApp() *fiber.App {
	return fiber.New(fiber.Config{BodyLimit: 4 * 1024})
}

var Module = fx.Options(
	fx.Provide(NewApp),
	fx.Provide(NewLinkHandler),
	fx.Provide(NewHandler),
	fx.Invoke(RegisterRoutes),
)
