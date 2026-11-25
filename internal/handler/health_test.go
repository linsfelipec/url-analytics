package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {

	request, err := http.NewRequest(
		http.MethodGet,
		"/healthz",
		nil,
	)
	if err != nil {
		t.Errorf("error creating request")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", Healthz)

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

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
