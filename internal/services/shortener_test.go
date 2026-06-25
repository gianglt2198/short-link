package services_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/gianglt1/short-link/internal/domain"
	helpermocks "github.com/gianglt1/short-link/internal/helpers/mocks"
	"github.com/gianglt1/short-link/internal/infra/logging"
	repomocks "github.com/gianglt1/short-link/internal/repositories/mocks"
	"github.com/gianglt1/short-link/internal/services"
)

func nopLogger() *logging.Logger { return &logging.Logger{Logger: zap.NewNop()} }

func TestEncode_ReturnsCode(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByURL", mock.Anything, "https://example.com/some/long/path").Return(domain.Link{}, domain.ErrLinkNotFound)
	gen.On("Generate").Return("ABC123", nil)
	repo.On("Save", mock.Anything, domain.Link{Code: "ABC123", OriginalURL: "https://example.com/some/long/path"}).Return(nil)

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	link, err := svc.Encode(t.Context(), "https://example.com/some/long/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.Code != "ABC123" {
		t.Errorf("expected code ABC123, got %q", link.Code)
	}
}

func TestEncode_Idempotent(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	existing := domain.Link{Code: "AAAAAA", OriginalURL: "https://idempotent.example.com/path"}

	repo.On("FindByURL", mock.Anything, "https://idempotent.example.com/path").
		Return(domain.Link{}, domain.ErrLinkNotFound).Once()
	gen.On("Generate").Return("AAAAAA", nil)
	repo.On("Save", mock.Anything, existing).Return(nil)
	repo.On("FindByURL", mock.Anything, "https://idempotent.example.com/path").
		Return(existing, nil).Once()

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	first, err := svc.Encode(t.Context(), "https://idempotent.example.com/path")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Encode(t.Context(), "https://idempotent.example.com/path")
	if err != nil {
		t.Fatal(err)
	}
	if first.Code != second.Code {
		t.Errorf("expected same code on second encode; got %q then %q", first.Code, second.Code)
	}
}

func TestEncode_DifferentURLsDifferentCodes(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByURL", mock.Anything, "https://aaa.example.com").Return(domain.Link{}, domain.ErrLinkNotFound)
	repo.On("FindByURL", mock.Anything, "https://bbb.example.com").Return(domain.Link{}, domain.ErrLinkNotFound)
	gen.On("Generate").Return("AAAAAA", nil).Once()
	gen.On("Generate").Return("BBBBBB", nil).Once()
	repo.On("Save", mock.Anything, mock.AnythingOfType("domain.Link")).Return(nil)

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	a, _ := svc.Encode(t.Context(), "https://aaa.example.com")
	b, _ := svc.Encode(t.Context(), "https://bbb.example.com")
	if a.Code == b.Code {
		t.Error("different URLs should produce different codes")
	}
}

func TestEncode_RetriesOnCollision(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByURL", mock.Anything, "https://b.example.com").Return(domain.Link{}, domain.ErrLinkNotFound)
	gen.On("Generate").Return("DUP000", nil).Once()
	repo.On("Save", mock.Anything, domain.Link{Code: "DUP000", OriginalURL: "https://b.example.com"}).Return(domain.ErrCodeCollision)
	gen.On("Generate").Return("FRESH1", nil).Once()
	repo.On("Save", mock.Anything, domain.Link{Code: "FRESH1", OriginalURL: "https://b.example.com"}).Return(nil)

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	link, err := svc.Encode(t.Context(), "https://b.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.Code != "FRESH1" {
		t.Errorf("expected retry to land on FRESH1, got %q", link.Code)
	}
}

func TestDecode_KnownCode(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	expected := domain.Link{Code: "CODE01", OriginalURL: "https://decode.example.com/path"}
	repo.On("FindByCode", mock.Anything, "CODE01").Return(expected, nil)

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	got, err := svc.Decode(t.Context(), "CODE01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OriginalURL != expected.OriginalURL {
		t.Errorf("got %q, want %q", got.OriginalURL, expected.OriginalURL)
	}
}

func TestDecode_UnknownCode(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	repo.On("FindByCode", mock.Anything, "zzzzzz").Return(domain.Link{}, domain.ErrLinkNotFound)

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	_, err := svc.Decode(t.Context(), "zzzzzz")
	if !errors.Is(err, domain.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

func TestEncode_InvalidScheme(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	_, err := svc.Encode(t.Context(), "ftp://example.com")
	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL for ftp scheme, got %v", err)
	}
}

func TestEncode_MissingScheme(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	_, err := svc.Encode(t.Context(), "example.com/path")
	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL for missing scheme, got %v", err)
	}
}

func TestEncode_URLTooLong(t *testing.T) {
	repo := repomocks.NewLinkRepository(t)
	gen := helpermocks.NewCodeGenerator(t)

	svc := services.NewShortenerService(services.ShortenerServiceParams{
		Repo:   repo,
		Gen:    gen,
		Logger: nopLogger(),
	})
	long := "https://example.com/" + strings.Repeat("a", 2048)
	_, err := svc.Encode(t.Context(), long)
	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL for over-long URL, got %v", err)
	}
}
