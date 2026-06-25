package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/domain"
	"github.com/gianglt1/short-link/internal/helpers"
	"github.com/gianglt1/short-link/internal/infra/cache"
	"github.com/gianglt1/short-link/internal/infra/logging"
	"github.com/gianglt1/short-link/internal/repositories"
	"github.com/gianglt1/short-link/internal/utils"
)

const maxRetries = 3

type (
	shortenerService struct {
		cfg *config.Config

		repo   repositories.LinkRepository
		gen    helpers.CodeGenerator
		logger *logging.Logger
		cache  cache.Cache
	}

	ShortenerService interface {
		Encode(ctx context.Context, rawURL string) (domain.Link, error)
		Decode(ctx context.Context, code string) (domain.Link, error)
	}
)

type ShortenerServiceParams struct {
	fx.In

	Repo   repositories.LinkRepository
	Gen    helpers.CodeGenerator
	Logger *logging.Logger
	Cache  cache.Cache
	Cfg    *config.Config
}

func NewShortenerService(params ShortenerServiceParams) ShortenerService {
	return &shortenerService{
		cfg:    params.Cfg,
		repo:   params.Repo,
		gen:    params.Gen,
		logger: params.Logger,
		cache:  params.Cache,
	}
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
		link := domain.Link{Code: code, OriginalURL: rawURL}
		if err := s.repo.Save(ctx, link); err != nil {
			if errors.Is(err, domain.ErrCodeCollision) {
				s.logger.GetWrappedLogger(ctx).Warn("encode: code collision, retrying", zap.String("code", code), zap.Int("attempt", i+1))
				continue
			}
			return domain.Link{}, err
		}
		return link, nil
	}

	return domain.Link{}, fmt.Errorf("failed to generate unique code after %d retries", maxRetries)
}

func (s *shortenerService) Decode(ctx context.Context, code string) (domain.Link, error) {
	if s.cache != nil {
		if val, err := s.cache.Get(ctx, code); err == nil && val != nil {
			return domain.Link{Code: code, OriginalURL: string(val)}, nil
		}
	}

	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return domain.Link{}, err
	}

	if s.cache != nil {
		if err := s.cache.Set(ctx, code, []byte(link.OriginalURL), time.Duration(s.cfg.Cache.TTL)*time.Second); err != nil {
			s.logger.GetWrappedLogger(ctx).Warn("decode: cache set failed", zap.String("code", code), zap.Error(err))
		}
	}

	return link, nil
}
