package handlers

import (
	"bytes"
	"errors"
	"html/template"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/domain"
	"github.com/gianglt1/short-link/internal/infra/logging"
	"github.com/gianglt1/short-link/internal/services"
	"github.com/gianglt1/short-link/internal/utils"
)

var (
	shortenTmpl = template.Must(template.New("shorten").Parse(
		`<div class="result-card result-success">` +
			`<p class="result-label">Your short link</p>` +
			`<div class="result-row">` +
			`<a class="result-url" href="{{.ShortURL}}" target="_blank" rel="noopener noreferrer">{{.ShortURL}}</a>` +
			`<button class="btn-copy" data-url="{{.ShortURL}}">Copy</button>` +
			`</div></div>`,
	))

	expandTmpl = template.Must(template.New("expand").Parse(
		`<div class="result-card result-success">` +
			`<p class="result-label">Original URL</p>` +
			`<div class="result-row">` +
			`<a class="result-url" href="{{.URL}}" target="_blank" rel="noopener noreferrer">{{.URL}}</a>` +
			`<button class="btn-open" data-url="{{.URL}}">Open</button>` +
			`</div></div>`,
	))

	errTmpl = template.Must(template.New("error").Parse(
		`<div class="result-card result-error">` +
			`<p class="result-message">{{.Message}}</p>` +
			`</div>`,
	))
)

type PageHandler struct {
	svc    services.ShortenerService
	cfg    *config.Config
	logger *logging.Logger
}

type PageHandlerParams struct {
	fx.In

	Svc    services.ShortenerService
	Cfg    *config.Config
	Logger *logging.Logger
}

func NewPageHandler(p PageHandlerParams) *PageHandler {
	return &PageHandler{svc: p.Svc, cfg: p.Cfg, logger: p.Logger}
}

func (h *PageHandler) Shorten(c *fiber.Ctx) error {
	if c.Get("HX-Request") != "true" {
		return c.Status(fiber.StatusBadRequest).SendString("htmx requests only")
	}

	ctx := c.Context()
	rawURL := c.FormValue("url")
	if rawURL == "" {
		return renderFrag(c, errTmpl, struct{ Message string }{"Enter a URL to shorten."})
	}

	link, err := h.svc.Encode(ctx, rawURL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			h.logger.GetWrappedLogger(ctx).Warn("shorten: invalid URL", zap.String("url", rawURL), zap.Error(err))
			return renderFrag(c, errTmpl, struct{ Message string }{"That doesn't look like a valid URL. Use http:// or https://."})
		}
		h.logger.GetWrappedLogger(ctx).Error("shorten: internal error", zap.Error(err))
		return renderFrag(c, errTmpl, struct{ Message string }{"Something went wrong. Try again."})
	}

	shortURL := h.cfg.Server.BaseURL + "/" + link.Code
	h.logger.GetWrappedLogger(ctx).Info("shorten: success", zap.String("code", link.Code))
	return renderFrag(c, shortenTmpl, struct{ ShortURL string }{shortURL})
}

func (h *PageHandler) Expand(c *fiber.Ctx) error {
	if c.Get("HX-Request") != "true" {
		return c.Status(fiber.StatusBadRequest).SendString("htmx requests only")
	}

	ctx := c.Context()
	rawShortURL := c.FormValue("short_url")
	if rawShortURL == "" {
		return renderFrag(c, errTmpl, struct{ Message string }{"Enter a short URL or code to expand."})
	}

	code := utils.ExtractCode(rawShortURL, h.cfg.Server.BaseURL)
	link, err := h.svc.Decode(ctx, code)
	if err != nil {
		if errors.Is(err, domain.ErrLinkNotFound) {
			return renderFrag(c, errTmpl, struct{ Message string }{"Short URL not found."})
		}
		h.logger.GetWrappedLogger(ctx).Error("expand: internal error", zap.Error(err))
		return renderFrag(c, errTmpl, struct{ Message string }{"Something went wrong. Try again."})
	}

	h.logger.GetWrappedLogger(ctx).Info("expand: success", zap.String("code", code))
	return renderFrag(c, expandTmpl, struct{ URL string }{link.OriginalURL})
}

func renderFrag(c *fiber.Ctx, tmpl *template.Template, data any) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("render error")
	}
	c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	return c.Status(fiber.StatusOK).Send(buf.Bytes())
}
