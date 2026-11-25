package gin_handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linsfelipec/url-analytics/internal/handler"
	"github.com/linsfelipec/url-analytics/internal/url"
)

type fakeURLService struct {
	createCalled    bool
	getByCodeCalled bool
	createFunc      func(ctx context.Context, originalURL string) (*url.URL, error)
	getByCodeFunc   func(ctx context.Context, code string) (*url.URL, error)
	validateUrlFunc func(rawUrl string) error
}

const BASE_URL_ADDRESS = "http://localhost:8080"

func (f *fakeURLService) Create(ctx context.Context, originalURL string) (*url.URL, error) {
	f.createCalled = true
	return f.createFunc(ctx, originalURL)
}

func (f *fakeURLService) GetByCode(ctx context.Context, code string) (*url.URL, error) {
	f.getByCodeCalled = true
	return f.getByCodeFunc(ctx, code)
}

func (f *fakeURLService) ValidateURL(rawUrl string) error {
	if f.validateUrlFunc == nil {
		return nil
	}
	return f.validateUrlFunc(rawUrl)
}

func TestHandler_InvalidBaseAddress(t *testing.T) {
	service := &fakeURLService{
		validateUrlFunc: func(rawUrl string) error {
			return url.ErrInvalidURL
		}}
	h, err := NewHandler(service, "http://Invalid")
	if !errors.Is(err, url.ErrInvalidURL) {
		t.Errorf("not validated base url")
	}
	if h != nil {
		t.Errorf("should not create handler with invalid base url")
	}
}

func testCreateSetup(t *testing.T, service *fakeURLService, recorder http.ResponseWriter, rBody string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(rBody),
	)
	request.Header.Set("Content-Type", "application/json")

	h, _ := NewHandler(service, BASE_URL_ADDRESS)

	router := gin.New()
	router.POST("/api/urls", h.Create)
	router.ServeHTTP(recorder, request)
}

func TestHandler_Create_Success(t *testing.T) {
	expectedURL := &url.URL{
		Code:        "abc123",
		OriginalURL: "https://example.com",
	}

	service := &fakeURLService{
		createFunc: func(ctx context.Context, originalURL string) (*url.URL, error) {
			if originalURL != expectedURL.OriginalURL {
				t.Errorf("expected service input %q, got %q",
					expectedURL.OriginalURL, originalURL)
			}
			return expectedURL, nil
		},
	}

	recorder := httptest.NewRecorder()
	testCreateSetup(t, service, recorder, fmt.Sprintf(`{"url":%q}`, expectedURL.OriginalURL))

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

	if response.ShortURL != (BASE_URL_ADDRESS + "/abc123") {
		t.Errorf("unexpected short URL: %q", response.ShortURL)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {

	service := fakeURLService{}
	recorder := httptest.NewRecorder()
	testCreateSetup(t, &service, recorder, `{"url":}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d",
			http.StatusBadRequest, recorder.Code)
	}
	if service.createCalled {
		t.Errorf("create function should NOT be called on invalid JSON")
	}
}

func TestHandler_Create_InvalidURL(t *testing.T) {

	service := fakeURLService{
		createFunc: func(ctx context.Context, originalURL string) (*url.URL, error) {
			return nil, url.ErrInvalidURL
		},
	}
	recorder := httptest.NewRecorder()
	testCreateSetup(t, &service, recorder, fmt.Sprintf(`{"url":%q}`, "ftp://example.com"))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d",
			http.StatusBadRequest, recorder.Code)
	}
	if !service.createCalled {
		t.Errorf("create function should be called")
	}
}

func TestHandler_Create_InternalError(t *testing.T) {

	service := fakeURLService{
		createFunc: func(ctx context.Context, originalURL string) (*url.URL, error) {
			return nil, errors.New("database connection failed")
		},
	}
	recorder := httptest.NewRecorder()
	testCreateSetup(t, &service, recorder, fmt.Sprintf(`{"url":%q}`, "ftp://example.com"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d",
			http.StatusInternalServerError, recorder.Code)
	}
	if !service.createCalled {
		t.Errorf("create function should be called")
	}

	if strings.Contains(recorder.Body.String(), "database connection failed") {
		t.Fatal("internal error details should not be exposed")
	}
}

func testGetByCodeSetup(t *testing.T, service *fakeURLService, recorder *httptest.ResponseRecorder, code string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/urls/%s", code),
		nil,
	)

	h, _ := NewHandler(service, BASE_URL_ADDRESS)
	router := gin.New()
	router.GET("/api/urls/:code", h.GetByCode)
	router.ServeHTTP(recorder, request)

}

func TestHandler_GetByCode_Success(t *testing.T) {
	expectedURL := &url.URL{
		Code:        "abc123",
		OriginalURL: "https://example.com",
	}

	service := &fakeURLService{
		getByCodeFunc: func(ctx context.Context, code string) (*url.URL, error) {
			if code != expectedURL.Code {
				t.Errorf("expected code %s, got %s", expectedURL.Code, code)
			}
			return expectedURL, nil
		},
	}

	recorder := httptest.NewRecorder()
	testGetByCodeSetup(t, service, recorder, expectedURL.Code)

	var response handler.GetURLResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if !service.getByCodeCalled {
		t.Fatalf("handler function not called: GetByCode")
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if response.Code != expectedURL.Code {
		t.Errorf("expected code %q, got %q", expectedURL.Code, response.Code)
	}

	if response.OriginalURL != expectedURL.OriginalURL {
		t.Errorf("expected original URL %q, got %q",
			expectedURL.OriginalURL, response.OriginalURL)
	}
}

func TestHandler_GetByCode_NotFound(t *testing.T) {
	service := &fakeURLService{
		getByCodeFunc: func(ctx context.Context, code string) (*url.URL, error) {
			return nil, url.ErrURLNotFound
		},
	}

	recorder := httptest.NewRecorder()
	testGetByCodeSetup(t, service, recorder, "unknown")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d",
			http.StatusNotFound, recorder.Code)
	}
}

func TestHandler_GetByCode_InternalError(t *testing.T) {
	service := &fakeURLService{
		getByCodeFunc: func(ctx context.Context, code string) (*url.URL, error) {
			return nil, errors.New("database connection failed")
		},
	}

	recorder := httptest.NewRecorder()

	testGetByCodeSetup(t, service, recorder, "abc123")
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d",
			http.StatusInternalServerError, recorder.Code)
	}

	if strings.Contains(recorder.Body.String(), "database connection failed") {
		t.Fatal("internal error details should not be exposed")
	}
}

func testRedirectSetup(t *testing.T, service *fakeURLService, recorder *httptest.ResponseRecorder, code string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/%s", code),
		nil,
	)

	h, _ := NewHandler(service, BASE_URL_ADDRESS)
	router := gin.New()
	router.GET("/:code", h.Redirect)
	router.ServeHTTP(recorder, request)
}

func TestHandler_Redirect_Success(t *testing.T) {
	expectedURL := &url.URL{
		Code:        "abc123",
		OriginalURL: "https://example.com",
	}

	service := &fakeURLService{
		getByCodeFunc: func(ctx context.Context, code string) (*url.URL, error) {
			if code != expectedURL.Code {
				t.Errorf("expected code %s, got %s", expectedURL.Code, code)
			}
			return expectedURL, nil
		},
	}

	recorder := httptest.NewRecorder()
	testRedirectSetup(t, service, recorder, expectedURL.Code)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d",
			http.StatusFound, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if location != expectedURL.OriginalURL {
		t.Errorf("expected Location %q, got %q",
			expectedURL.OriginalURL, location)
	}
}

func TestHandler_Redirect_NotFound(t *testing.T) {
	service := &fakeURLService{
		getByCodeFunc: func(ctx context.Context, code string) (*url.URL, error) {
			return nil, url.ErrURLNotFound
		},
	}

	recorder := httptest.NewRecorder()
	testRedirectSetup(t, service, recorder, "unknown")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d",
			http.StatusNotFound, recorder.Code)
	}
}

func TestHandler_Redirect_InternalError(t *testing.T) {
	service := &fakeURLService{
		getByCodeFunc: func(ctx context.Context, code string) (*url.URL, error) {
			return nil, errors.New("database connection failed")
		},
	}

	recorder := httptest.NewRecorder()
	testRedirectSetup(t, service, recorder, "abc123")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d",
			http.StatusInternalServerError, recorder.Code)
	}

	if strings.Contains(recorder.Body.String(), "database connection failed") {
		t.Fatal("internal error details should not be exposed")
	}
}
