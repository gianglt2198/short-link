package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/domain"
	"github.com/gianglt1/short-link/internal/dtos"
	"github.com/gianglt1/short-link/internal/infra/logging"
	"github.com/gianglt1/short-link/internal/services"
	"github.com/gianglt1/short-link/internal/utils"
)

type LinkHandler struct {
	svc    services.ShortenerService
	cfg    *config.Config
	logger *logging.Logger
}

type LinkHandlerParams struct {
	fx.In

	Svc    services.ShortenerService
	Cfg    *config.Config
	Logger *logging.Logger
}

func NewLinkHandler(p LinkHandlerParams) *LinkHandler {
	return &LinkHandler{svc: p.Svc, cfg: p.Cfg, logger: p.Logger}
}

func (h *LinkHandler) Encode(c *fiber.Ctx) error {
	ctx := c.Context()

	var req dtos.EncodeRequest
	if err := c.BodyParser(&req); err != nil || req.URL == "" {
		h.logger.GetWrappedLogger(ctx).Warn("encode: invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	link, err := h.svc.Encode(ctx, req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			h.logger.GetWrappedLogger(ctx).Warn("encode: invalid URL", zap.String("url", req.URL), zap.Error(err))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		h.logger.GetWrappedLogger(ctx).Error("encode: internal error", zap.String("url", req.URL), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	shortURL := h.cfg.Server.BaseURL + "/" + link.Code
	h.logger.GetWrappedLogger(ctx).Info("encode: success", zap.String("code", link.Code), zap.String("short_url", shortURL))
	return c.JSON(dtos.EncodeResponse{ShortURL: shortURL})
}

func (h *LinkHandler) Decode(c *fiber.Ctx) error {
	ctx := c.Context()

	var req dtos.DecodeRequest
	if err := c.BodyParser(&req); err != nil || req.ShortURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	code := utils.ExtractCode(req.ShortURL, h.cfg.Server.BaseURL)
	if !utils.IsValidCode(code) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short URL not found"})
	}
	h.logger.GetWrappedLogger(ctx).Info("decode: request received", zap.String("code", code))

	link, err := h.svc.Decode(ctx, code)
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
	ctx := c.Context()
	code := c.Params("code")

	if !utils.IsValidCode(code) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short URL not found"})
	}
	h.logger.GetWrappedLogger(ctx).Info("redirect: request received", zap.String("code", code))

	link, err := h.svc.Decode(ctx, code)
	if err != nil {
		if errors.Is(err, domain.ErrLinkNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short URL not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	return c.Redirect(link.OriginalURL, fiber.StatusFound)
}
