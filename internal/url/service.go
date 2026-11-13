package url

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
)

type URLRepository interface {
	Create(ctx context.Context, code, originalURL string) (*URL, error)
	GetByCode(ctx context.Context, code string) (*URL, error)
}

type Service struct {
	repo URLRepository
}

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

	buffer := make([]byte, 6)
	rand.Read(buffer)
	code := hex.EncodeToString(buffer)
	return s.repo.Create(ctx, code, originalURL)
}
