package url

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgconn"
)

type URLRepository interface {
	Create(ctx context.Context, code, originalURL string) (*URL, error)
	GetByCode(ctx context.Context, code string) (*URL, error)
}

type Service struct {
	repo URLRepository
}

const maxCodeGenerationAttempts = 3

func NewService(repo URLRepository) *Service {
	return &Service{repo: repo}
}

func validateURL(rawURL string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %v", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain a Host name")
	}

	return nil
}

func (s *Service) Create(ctx context.Context, originalURL string) (*URL, error) {

	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	for range maxCodeGenerationAttempts {
		buffer := make([]byte, 6)
		rand.Read(buffer)
		code := hex.EncodeToString(buffer)
		createdURL, err := s.repo.Create(ctx, code, originalURL)
		if err == nil {
			return createdURL, nil
		} else if !isDuplicateCodeError(err) {
			return nil, fmt.Errorf("service creating failure: %w", err)
		}
	}

	return nil, fmt.Errorf("failed to generate a unique code after %d attempts", maxCodeGenerationAttempts)
}

func isDuplicateCodeError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.ConstraintName == "urls_code_key"
}
