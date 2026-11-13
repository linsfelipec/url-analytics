package url

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, code, originalURL string) (*URL, error) {
	query := `
		INSERT INTO urls (code, original_url)
		VALUES ($1, $2)
	RETURNING id, code, original_url, created_at
	`

	var u URL
	err := r.db.QueryRowContext(ctx, query, code, originalURL).
		Scan(&u.ID, &u.Code, &u.OriginalURL, &u.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert URL: %w", err)
	}

	return &u, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*URL, error) {
	var u URL
	query := `
		SELECT id, code, original_url, created_at
		FROM urls
		WHERE code = $1
	`
	if err := r.db.QueryRowContext(ctx, query, code).Scan(
		&u.ID, &u.Code, &u.OriginalURL, &u.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("getting URL by code: %w", err)
	}

	return &u, nil
}
