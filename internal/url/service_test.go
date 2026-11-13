package url

import (
	"context"
	"encoding/hex"
	"testing"
)

func TestValidateURL(t *testing.T) {
	test := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid HTTPS URL",
			input:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL with path",
			input:   "http://example.com/products",
			wantErr: false,
		},
		{
			name:    "unsupported FTP scheme",
			input:   "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "unsupported file scheme",
			input:   "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			input:   "example.com",
			wantErr: true,
		},
		{
			name:    "missing hostname",
			input:   "https:///products",
			wantErr: true,
		},
		{
			name:    "empty URL",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateURL(tt.input); (err == nil) == tt.wantErr {
				t.Errorf("validate URL(%q) error: %v, wantErr: %v", tt.input, err, tt.wantErr)
			}

		})
	}
}

type fakeURLRepository struct {
	createCalled bool
	createdCode  string
	createdURL   string
}

func (f *fakeURLRepository) Create(ctx context.Context, code, originalURL string) (*URL, error) {
	f.createCalled = true
	f.createdCode = code
	f.createdURL = originalURL

	return &URL{
		Code:        code,
		OriginalURL: originalURL,
	}, nil
}

func (f *fakeURLRepository) GetByCode(
	ctx context.Context,
	code string,
) (*URL, error) {
	return nil, nil
}

func TestService_Create(t *testing.T) {
	repo := &fakeURLRepository{}
	service := NewService(repo)

	originalURL := "https://example.com/products"

	created, err := service.Create(context.Background(), originalURL)

	if err != nil {
		t.Fatalf("Create() returned an error: %v", err)
	}

	if !repo.createCalled {
		t.Fatal("expected repository Create() to be called")
	}

	if repo.createdURL != originalURL {
		t.Errorf("expected URL %q, got %q",
			originalURL, repo.createdURL)
	}

	if created.Code == "" {
		t.Error("expected generated short code, got empty string")
	}

	if len(created.Code) != 12 {
		t.Errorf("expected code length 12, got %d", len(created.Code))
	}

	if _, err := hex.DecodeString(created.Code); err != nil {
		t.Errorf("expected hexadecimal code, got %q", created.Code)
	}
}

func TestService_Create_InvalidURL(t *testing.T) {
	repo := &fakeURLRepository{}
	service := NewService(repo)

	_, err := service.Create(context.Background(), "ftp://example.com")

	if err == nil {
		t.Fatal("expected an error for unsupported URL scheme")
	}

	if repo.createCalled {
		t.Fatal("repository Create() should not be called for an invalid URL")
	}
}
