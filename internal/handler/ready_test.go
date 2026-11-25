package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeReadyRepository struct {
	pingFunc func(context.Context) error
}

func (f *fakeReadyRepository) PingContext(ctx context.Context) error {
	return f.pingFunc(ctx)
}

func setupTest(t *testing.T, repo ReadyRepository) *httptest.ResponseRecorder {
	t.Helper()

	h := Readyz(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", h)

	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)
	return recorder
}

func TestReadyz_Success(t *testing.T) {
	repo := fakeReadyRepository{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}
	recorder := setupTest(t, &repo)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected code %d, got %d", http.StatusOK, recorder.Code)
	}

	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected application/json, got %s", got)
	}

	var response StatusResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected ok, got: %s", response.Status)
	}
}

func TestReadyz_DatabaseUnavailable(t *testing.T) {
	repo := fakeReadyRepository{
		pingFunc: func(ctx context.Context) error {
			return sql.ErrConnDone
		},
	}
	recorder := setupTest(t, &repo)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("expected code %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected application/json, got %s", got)
	}

	var response StatusResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if response.Status != "unavailable" {
		t.Errorf("expected unavailable, got: %s", response.Status)
	}
}
