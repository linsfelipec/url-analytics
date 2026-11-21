package std_handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/linsfelipec/url-analytics/internal/handler"
	"github.com/linsfelipec/url-analytics/internal/url"
)

type fakeURLService struct {
	createFunc func(ctx context.Context, originalURL string) (*url.URL, error)
}

func (f *fakeURLService) Create(ctx context.Context, originalURL string) (*url.URL, error) {
	return f.createFunc(ctx, originalURL)
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	service := fakeURLService{}
	h := NewHandler(&service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":}`),
	)
	recorder := httptest.NewRecorder()

	h.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d",
			http.StatusBadRequest, recorder.Code)
	}
}

func TestHandler_Create_InvalidURL(t *testing.T) {
	service := &fakeURLService{
		createFunc: func(ctx context.Context, originalURL string) (*url.URL, error) {
			return nil, url.ErrInvalidURL
		},
	}

	h := NewHandler(service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"ftp://example.com"}`),
	)
	recorder := httptest.NewRecorder()

	h.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d",
			http.StatusBadRequest, recorder.Code)
	}
}

func TestHandler_Create_InternalError(t *testing.T) {
	service := &fakeURLService{
		createFunc: func(ctx context.Context, originalURL string) (*url.URL, error) {
			return nil, errors.New("random internal error")
		},
	}

	h := NewHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.com"}`),
	)

	recorder := httptest.NewRecorder()
	h.Create(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status: %d, received status %d",
			http.StatusInternalServerError,
			recorder.Code)
	}

	if strings.Contains(recorder.Body.String(), "random internal error") {
		t.Fatal("internal error details should not be exposed")
	}
}

func TestHandler_Create_Success(t *testing.T) {
	expectedURL := &url.URL{
		Code:        "abc123",
		OriginalURL: "https://example.com",
	}

	service := &fakeURLService{
		createFunc: func(ctx context.Context, originalURL string) (*url.URL, error) {
			return expectedURL, nil
		},
	}

	h := NewHandler(service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.com"}`),
	)
	recorder := httptest.NewRecorder()

	h.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d",
			http.StatusCreated, recorder.Code)
	}

	var response handler.CreateURLResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if response.Code != expectedURL.Code {
		t.Errorf("expected code %q, got %q", expectedURL.Code, response.Code)
	}

	if response.OriginalURL != expectedURL.OriginalURL {
		t.Errorf("expected original URL %q, got %q",
			expectedURL.OriginalURL, response.OriginalURL)
	}

	if response.ShortURL != "http://localhost:8080/abc123" {
		t.Errorf("unexpected short URL: %q", response.ShortURL)
	}
}

func TestHandler_Create_Route(t *testing.T) {
	expectedUrl := &url.URL{
		Code:        "abc123",
		OriginalURL: "https://example.test",
	}

	service := &fakeURLService{
		createFunc: func(ctx context.Context, originalURL string) (*url.URL, error) {
			return expectedUrl, nil
		},
	}

	h := NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/urls", h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.com"}`),
	)

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status: %d, received: %d",
			http.StatusCreated,
			recorder.Code)
	}
}

func TestHandler_Create_RouteWrongMethod(t *testing.T) {
	service := &fakeURLService{}
	h := NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/urls", h.Create)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/urls",
		nil,
	)

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d",
			http.StatusMethodNotAllowed, recorder.Code)
	}
}
