package repositories_test

import (
	"errors"
	"os"
	"testing"

	"go.uber.org/fx"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/domain"
	"github.com/gianglt1/short-link/internal/infra/database"
	"github.com/gianglt1/short-link/internal/repositories"
)

func newTestRepo(t *testing.T) repositories.LinkRepository {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set; skipping postgres integration tests")
	}
	cfg := &config.Config{
		Database: config.DatabaseConfig{URL: dbURL},
		Server:   config.ServerConfig{BaseURL: "http://localhost:8080"},
	}

	var repo repositories.LinkRepository
	app := fx.New(
		fx.Supply(cfg),
		fx.Provide(database.NewPool),
		repositories.Module,
		fx.Populate(&repo),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx.New: %v", err)
	}
	t.Cleanup(func() { _ = app.Stop(t.Context()) })
	if err := app.Start(t.Context()); err != nil {
		t.Fatalf("app.Start: %v", err)
	}
	return repo
}

func TestPgRepository_SaveAndFindByCode(t *testing.T) {
	repo := newTestRepo(t)
	ctx := t.Context()
	if err := repo.Save(ctx, domain.Link{Code: "abc123", OriginalURL: "https://example.com"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.FindByCode(ctx, "abc123")
	if err != nil {
		t.Fatalf("FindByCode: %v", err)
	}
	if got.OriginalURL != "https://example.com" {
		t.Errorf("got %q, want %q", got.OriginalURL, "https://example.com")
	}
}

func TestPgRepository_FindByURL(t *testing.T) {
	repo := newTestRepo(t)
	ctx := t.Context()
	_ = repo.Save(ctx, domain.Link{Code: "xyz999", OriginalURL: "https://findbyurl.example.com"})
	got, err := repo.FindByURL(ctx, "https://findbyurl.example.com")
	if err != nil {
		t.Fatalf("FindByURL: %v", err)
	}
	if got.Code != "xyz999" {
		t.Errorf("got %q, want %q", got.Code, "xyz999")
	}
}

func TestPgRepository_FindByCode_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	_, err := repo.FindByCode(t.Context(), "zzzzzz")
	if err != domain.ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

func TestPgRepository_Save_CodeCollision(t *testing.T) {
	repo := newTestRepo(t)
	ctx := t.Context()
	_ = repo.Save(ctx, domain.Link{Code: "col123", OriginalURL: "https://collision1.example.com"})
	err := repo.Save(ctx, domain.Link{Code: "col123", OriginalURL: "https://collision2.example.com"})
	if !errors.Is(err, domain.ErrCodeCollision) {
		t.Errorf("expected ErrCodeCollision on duplicate code, got %v", err)
	}
}

func TestPgRepository_Save_DuplicateErrors(t *testing.T) {
	repo := newTestRepo(t)
	ctx := t.Context()
	link := domain.Link{Code: "idem01", OriginalURL: "https://idempotent.example.com"}
	_ = repo.Save(ctx, link)
	if err := repo.Save(ctx, link); err == nil {
		t.Error("expected error on duplicate save, got nil")
	}
}
