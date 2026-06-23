package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/gianglt1/short-link/internal/domain"
)

type (
	PgLinkRepository struct {
		pool *pgxpool.Pool
	}

	LinkRepository interface {
		Save(ctx context.Context, link domain.Link) error
		FindByCode(ctx context.Context, code string) (domain.Link, error)
		FindByURL(ctx context.Context, url string) (domain.Link, error)
		CodeExists(ctx context.Context, code string) (bool, error)
	}
)

type PgLinkRepositoryResult struct {
	fx.Out

	Repo LinkRepository
}

func NewPgLinkRepository(pool *pgxpool.Pool) PgLinkRepositoryResult {
	return PgLinkRepositoryResult{Repo: &PgLinkRepository{pool: pool}}
}

func (r *PgLinkRepository) Save(ctx context.Context, link domain.Link) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO links (code, original_url) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		link.Code, link.OriginalURL,
	)
	return err
}

func (r *PgLinkRepository) FindByCode(ctx context.Context, code string) (domain.Link, error) {
	link := domain.Link{Code: code}
	err := r.pool.QueryRow(ctx,
		`SELECT original_url, created_at FROM links WHERE code = $1`,
		code,
	).Scan(&link.OriginalURL, &link.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Link{}, domain.ErrLinkNotFound
	}
	if err != nil {
		return domain.Link{}, err
	}
	return link, nil
}

func (r *PgLinkRepository) FindByURL(ctx context.Context, originalURL string) (domain.Link, error) {
	link := domain.Link{OriginalURL: originalURL}
	err := r.pool.QueryRow(ctx,
		`SELECT code, created_at FROM links WHERE original_url = $1`,
		originalURL,
	).Scan(&link.Code, &link.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Link{}, domain.ErrLinkNotFound
	}
	if err != nil {
		return domain.Link{}, err
	}
	return link, nil
}

func (r *PgLinkRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM links WHERE code = $1)`,
		code,
	).Scan(&exists)
	return exists, err
}
