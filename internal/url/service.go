package url

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
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

var ErrURLNotFound = errors.New("URL not found")
var ErrInvalidURL = errors.New("invalid URL")

const maxCodeGenerationAttempts = 3

func NewService(repo URLRepository) *Service {
	return &Service{repo: repo}
}

func validateURL(rawURL string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: unsupported scheme %q", ErrInvalidURL, parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%w: missing host", ErrInvalidURL)
	}

	return nil
}

func (s *Service) Create(ctx context.Context, originalURL string) (*URL, error) {

	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	for range maxCodeGenerationAttempts {
		code, _ := generateCode()
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

func (s *Service) GetByCode(ctx context.Context, code string) (*URL, error) {
	url, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrURLNotFound
		}
		return nil, fmt.Errorf("getting URL by code: %w", err)
	}

	return url, nil
}

const (
	codeLength     = 6
	alphabetLength = 52
)

func generateCode() (string, error) {
	code := make([]byte, codeLength)

	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(alphabetLength)))
		if err != nil {
			return "", err
		}

		c := byte(n.Int64())
		if c < 26 {
			code[i] = 'A' + c
		} else {
			code[i] = 'a' + (c - 26)
		}
	}

	return string(code), nil
}
