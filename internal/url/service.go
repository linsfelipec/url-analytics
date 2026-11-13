package url

import (
	"context"
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

func ValidateURL(rawURL string) error {
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
