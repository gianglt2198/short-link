package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/domain"
	"github.com/gianglt1/short-link/internal/handlers"
	helpermocks "github.com/gianglt1/short-link/internal/helpers/mocks"
	"github.com/gianglt1/short-link/internal/infra/logging"
	repomocks "github.com/gianglt1/short-link/internal/repositories/mocks"
	"github.com/gianglt1/short-link/internal/services"
)

func newTestApp(t *testing.T, repo *repomocks.LinkRepository, gen *helpermocks.CodeGenerator) *fiber.App {
	cfg := &config.Config{Server: config.ServerConfig{BaseURL: "http://localhost:8080"}}
	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo: repo,
		Gen:  gen,
	})
	log := logging.NewNoopLogger()
	lh := handlers.NewLinkHandler(handlers.LinkHandlerParams{Svc: svc, Cfg: cfg, Logger: log})
	ph := handlers.NewPageHandler(handlers.PageHandlerParams{Svc: svc, Cfg: cfg, Logger: log})
	h := handlers.NewHandler(lh, ph)
	app := handlers.NewApp(logging.NewNoopLogger(), cfg)
	handlers.RegisterRoutes(app, h)
	return app
}

func post(app *fiber.App, path, body string) *http.Response {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	return resp
}

func TestEncode_ValidURL(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByURL", mock.Anything, "https://example.com/path").Return(domain.Link{}, domain.ErrLinkNotFound)
	gen.On("Generate").Return("STUB01", nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("domain.Link")).Return(nil)

	app := newTestApp(t, repo, gen)
	resp := post(app, "/encode", `{"url":"https://example.com/path"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if !strings.HasPrefix(body["short_url"], "http://localhost:8080/") {
		t.Errorf("unexpected short_url: %q", body["short_url"])
	}
}

func TestEncode_InvalidURL(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	app := newTestApp(t, repo, gen)
	resp := post(app, "/encode", `{"url":"not-a-url"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestEncode_MissingURLField(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	app := newTestApp(t, repo, gen)
	resp := post(app, "/encode", `{}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestEncode_InternalError(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByURL", mock.Anything, "https://example.com/path").Return(domain.Link{}, fmt.Errorf("db connection lost"))

	app := newTestApp(t, repo, gen)
	resp := post(app, "/encode", `{"url":"https://example.com/path"}`)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestDecode_KnownCode(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByURL", mock.Anything, "https://decode-test.example.com").Return(domain.Link{}, domain.ErrLinkNotFound)
	gen.On("Generate").Return("STUB01", nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("domain.Link")).Return(nil)
	repo.On("FindByCode", mock.Anything, "STUB01").Return(
		domain.Link{Code: "STUB01", OriginalURL: "https://decode-test.example.com"}, nil,
	)

	app := newTestApp(t, repo, gen)

	encResp := post(app, "/encode", `{"url":"https://decode-test.example.com"}`)
	var encBody map[string]string
	json.NewDecoder(encResp.Body).Decode(&encBody)

	body, _ := json.Marshal(map[string]string{"short_url": encBody["short_url"]})
	resp := post(app, "/decode", string(body))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var decBody map[string]string
	json.NewDecoder(resp.Body).Decode(&decBody)
	if decBody["url"] != "https://decode-test.example.com" {
		t.Errorf("got %q", decBody["url"])
	}
}

func TestDecode_UnknownCode(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByCode", mock.Anything, "zzzzzz").Return(domain.Link{}, domain.ErrLinkNotFound)

	app := newTestApp(t, repo, gen)
	resp := post(app, "/decode", `{"short_url":"http://localhost:8080/zzzzzz"}`)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDecode_InternalError(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByCode", mock.Anything, "zzzzzz").Return(domain.Link{}, fmt.Errorf("db connection lost"))

	app := newTestApp(t, repo, gen)
	resp := post(app, "/decode", `{"short_url":"http://localhost:8080/zzzzzz"}`)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestDecode_MalformedJSON(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	app := newTestApp(t, repo, gen)
	resp := post(app, "/decode", `{bad json`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestEncode_MalformedJSON(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	app := newTestApp(t, repo, gen)
	resp := post(app, "/encode", `{bad json`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func get(app *fiber.App, path string) *http.Response {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, _ := app.Test(req)
	return resp
}

func TestRedirect_KnownCode(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByCode", mock.Anything, "ABC123").Return(
		domain.Link{Code: "ABC123", OriginalURL: "https://example.com/long/path"}, nil,
	)

	app := newTestApp(t, repo, gen)
	resp := get(app, "/ABC123")
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 301, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "https://example.com/long/path" {
		t.Errorf("unexpected Location: %q", loc)
	}
}

func TestRedirect_UnknownCode(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByCode", mock.Anything, "zzzzzz").Return(domain.Link{}, domain.ErrLinkNotFound)

	app := newTestApp(t, repo, gen)
	resp := get(app, "/zzzzzz")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRedirect_InternalError(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByCode", mock.Anything, "zzzzzz").Return(domain.Link{}, fmt.Errorf("db connection lost"))

	app := newTestApp(t, repo, gen)
	resp := get(app, "/zzzzzz")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}
