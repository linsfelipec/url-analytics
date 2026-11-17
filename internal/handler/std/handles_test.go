package std_handler

import (
	"context"

	"github.com/linsfelipec/url-analytics/internal/url"
)

type fakeURLService struct {
	createFunc func(ctx context.Context, originalURL string) (*url.URL, error)
}

func (f *fakeURLService) Create(ctx context.Context, originalURL string) (*url.URL, error) {
	return f.createFunc(ctx, originalURL)
}
