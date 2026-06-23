package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/gianglt1/short-link/internal/domain"
	"github.com/gianglt1/short-link/internal/helpers"
	"github.com/gianglt1/short-link/internal/repositories"
	"github.com/gianglt1/short-link/internal/utils"
	"go.uber.org/fx"
)

const maxRetries = 3

type (
	shortenerService struct {
		repo repositories.LinkRepository
		gen  helpers.CodeGenerator
	}

	ShortenerService interface {
		Encode(ctx context.Context, rawURL string) (domain.Link, error)
		Decode(ctx context.Context, code string) (domain.Link, error)
	}
)

type ShortenerServiceParams struct {
	fx.In

	Repo repositories.LinkRepository
	Gen  helpers.CodeGenerator
}

func NewShortenerService(params ShortenerServiceParams) ShortenerService {
	return &shortenerService{repo: params.Repo, gen: params.Gen}
}

func (s *shortenerService) Encode(ctx context.Context, rawURL string) (domain.Link, error) {
	if err := utils.ValidateURL(rawURL); err != nil {
		return domain.Link{}, err
	}

	if link, err := s.repo.FindByURL(ctx, rawURL); err == nil {
		return link, nil
	} else if !errors.Is(err, domain.ErrLinkNotFound) {
		return domain.Link{}, err
	}

	for i := 0; i < maxRetries; i++ {
		code, err := s.gen.Generate()
		if err != nil {
			return domain.Link{}, fmt.Errorf("generate code: %w", err)
		}
		exists, err := s.repo.CodeExists(ctx, code)
		if err != nil {
			return domain.Link{}, err
		}
		if exists {
			continue
		}
		link := domain.Link{Code: code, OriginalURL: rawURL}
		if err := s.repo.Save(ctx, link); err != nil {
			return domain.Link{}, err
		}
		return link, nil
	}
	return domain.Link{}, fmt.Errorf("failed to generate unique code after %d retries", maxRetries)
}

func (s *shortenerService) Decode(ctx context.Context, code string) (domain.Link, error) {
	return s.repo.FindByCode(ctx, code)
}
