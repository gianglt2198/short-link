package handlers

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"go.uber.org/fx"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/infra/logging"
	"github.com/gianglt1/short-link/internal/middlewares"
)

//go:embed web
var webFS embed.FS

const cspHeader = "default-src 'self'; " +
	"script-src 'self' https://unpkg.com; " +
	"style-src 'self' https://fonts.googleapis.com; " +
	"font-src https://fonts.gstatic.com; " +
	"connect-src 'self'; " +
	"img-src 'self' data:; " +
	"frame-ancestors 'none'"

type Handler struct {
	link *LinkHandler
	page *PageHandler
}

func NewHandler(link *LinkHandler, page *PageHandler) *Handler {
	return &Handler{link: link, page: page}
}

func RegisterRoutes(app *fiber.App, h *Handler) {
	// Static files from embedded FS — calls Next() when file not found,
	// so /:code and API routes below still work correctly.
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(webFS),
		PathPrefix: "web",
		Index:      "index.html",
	}))

	// JSON API (unchanged)
	app.Post("/encode", h.link.Encode)
	app.Post("/decode", h.link.Decode)

	// HTMX fragment endpoints
	app.Post("/shorten", h.page.Shorten)
	app.Post("/expand", h.page.Expand)

	// Redirect — must be last to avoid catching API paths
	app.Get("/:code", h.link.Redirect)
}

func NewApp(log *logging.Logger, cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 4 * 1024})
	app.Use(middlewares.RequestIDMiddleware)
	app.Use(middlewares.LoggerMiddleware(log, cfg))
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Content-Security-Policy", cspHeader)
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-Content-Type-Options", "nosniff")
		return c.Next()
	})
	return app
}

var Module = fx.Options(
	fx.Provide(NewApp),
	fx.Provide(NewLinkHandler),
	fx.Provide(NewPageHandler),
	fx.Provide(NewHandler),
	fx.Invoke(RegisterRoutes),
)
