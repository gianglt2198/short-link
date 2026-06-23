package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/domain"
	"github.com/gianglt1/short-link/internal/dtos"
	"github.com/gianglt1/short-link/internal/services"
	"github.com/gianglt1/short-link/internal/utils"
)

type LinkHandler struct {
	svc services.ShortenerService
	cfg *config.Config
}

type LinkHandlerParams struct {
	fx.In

	Svc services.ShortenerService
	Cfg *config.Config
}

func NewLinkHandler(p LinkHandlerParams) *LinkHandler {
	return &LinkHandler{svc: p.Svc, cfg: p.Cfg}
}

func (h *LinkHandler) Encode(c *fiber.Ctx) error {
	var req dtos.EncodeRequest
	if err := c.BodyParser(&req); err != nil || req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	link, err := h.svc.Encode(c.Context(), req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	return c.JSON(dtos.EncodeResponse{ShortURL: h.cfg.Server.BaseURL + "/" + link.Code})
}

func (h *LinkHandler) Decode(c *fiber.Ctx) error {
	var req dtos.DecodeRequest
	if err := c.BodyParser(&req); err != nil || req.ShortURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	code := utils.ExtractCode(req.ShortURL, h.cfg.Server.BaseURL)
	link, err := h.svc.Decode(c.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrLinkNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short URL not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	return c.JSON(dtos.DecodeResponse{URL: link.OriginalURL})
}

// Redirect decodes the short code and redirects to the destination URL.
// Uses 302 to preserve analytics potential (request counts per URL).
func (h *LinkHandler) Redirect(c *fiber.Ctx) error {
	code := c.Params("code")
	link, err := h.svc.Decode(c.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrLinkNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short URL not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.Redirect(link.OriginalURL, fiber.StatusFound)
}
