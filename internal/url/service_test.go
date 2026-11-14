package url

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
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
	createCalls  int
	createdCode  string
	createdURL   string
	createFunc   func(code, originalURL string) (*URL, error)
}

func (f *fakeURLRepository) Create(
	ctx context.Context,
	code, originalURL string,
) (*URL, error) {
	f.createCalled = true
	f.createCalls++
	f.createdCode = code
	f.createdURL = originalURL

	if f.createFunc != nil {
		return f.createFunc(code, originalURL)
	}

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

func TestService_Create_RetryOnDuplicateCode(t *testing.T) {
	repo := &fakeURLRepository{}
	service := NewService(repo)

	repo.createFunc = func(code string, originalURL string) (*URL, error) {
		if repo.createCalls < 2 {
			return nil, &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "urls_code_key",
			}
		}
		return &URL{
			Code:        code,
			OriginalURL: originalURL,
		}, nil
	}

	createdURL, err := service.Create(context.Background(), "https://example.com/products")

	if err != nil {
		t.Fatalf("Create() returned an error: %v", err)
	}

	if repo.createCalls != 2 {
		t.Errorf("expected 2 repository calls, got %d", repo.createCalls)
	}

	if createdURL == nil || createdURL.Code == "" {
		t.Fatal("expected a created URL with a short code")
	}
}

func TestService_Create_ExceedMaxAttempts(t *testing.T) {
	repo := &fakeURLRepository{}
	s := NewService(repo)

	repo.createFunc = func(code string, originalURL string) (*URL, error) {
		return nil, &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "urls_code_key",
		}
	}

	createdURL, err := s.Create(context.Background(), "https://example.com/products")

	if err == nil {
		t.Fatalf("expected error after too many tries")
	}

	if repo.createCalls != maxCodeGenerationAttempts {
		t.Fatalf("expected %d attempts, got %d", maxCodeGenerationAttempts, repo.createCalls)
	}

	if createdURL != nil {
		t.Fatalf("expected URL not created, got %+v", createdURL)
	}
}

func TestService_Create_NoDuplicateError(t *testing.T) {
	repo := &fakeURLRepository{}
	s := NewService(repo)

	repo.createFunc = func(code string, originalURL string) (*URL, error) {
		return nil, errors.New("fake error")
	}

	createdURL, err := s.Create(context.Background(), "https://example.com/products")

	if err.Error() != "service creating failure: fake error" {
		t.Errorf("unexpected error: %v", err)
	}

	if repo.createCalls != 1 {
		t.Fatalf("expected 1 attempt, got %d", repo.createCalls)
	}

	if createdURL != nil {
		t.Fatalf("expected URL not created, got %+v", createdURL)
	}
}
