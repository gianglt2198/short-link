package middlewares

import (
	"regexp"

	"github.com/gofiber/fiber/v2"
)

var codePattern = regexp.MustCompile(`^[a-zA-Z0-9]{6}$`)

var allowedPOST = map[string]bool{
	"/encode":  true,
	"/decode":  true,
	"/shorten": true,
	"/expand":  true,
}

var allowedGET = map[string]bool{
	"/":        true,
	"/app.js":  true,
	"/app.css": true,
}

// AllowlistMiddleware blocks any request not on the known exposed surface.
// GET requests pass through to let the embedded filesystem middleware handle
// static assets; single-segment paths are validated as 6-char base62 codes
// to prevent scanner paths from reaching the redirect route.
func AllowlistMiddleware(c *fiber.Ctx) error {
	method := c.Method()
	path := c.Path()

	switch method {
	case fiber.MethodGet:
		if !allowedGET[path] && !codePattern.MatchString(path[1:]) {
			return c.SendStatus(fiber.StatusNotFound)
		}
	case fiber.MethodPost:
		if !allowedPOST[path] {
			return c.SendStatus(fiber.StatusNotFound)
		}
	default:
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.Next()
}
