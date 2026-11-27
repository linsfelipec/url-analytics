package std_handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linsfelipec/url-analytics/internal/logger"
)

type loggerMsg struct {
	Msg      string  `json:"msg"`
	Method   string  `json:"method"`
	Path     string  `json:"path"`
	Status   int     `json:"status"`
	Duration float64 `json:"duration"`
	ClientIP string  `json:"client_addr"`
}

func TestMiddleware_RequestLogger(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		handler http.HandlerFunc
	}{
		{
			name:   "write header",
			status: http.StatusCreated,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
		},
		{
			name:   "write call",
			status: http.StatusOK,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write(nil)
			},
		},
		{
			name:   "write call with explicit status",
			status: http.StatusInternalServerError,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(nil)
			},
		},
	}

	const (
		path   = "/cmd/api"
		method = http.MethodGet
		ip     = "127.0.0.1:54321"
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			testLogger := logger.New(&logBuffer)

			request, err := http.NewRequest(method, path, nil)
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}
			request.RemoteAddr = ip

			mux := http.NewServeMux()
			mux.HandleFunc(path, tt.handler)

			var response httptest.ResponseRecorder
			Chain(mux, RequestLogger(testLogger)).ServeHTTP(&response, request)

			if response.Code != tt.status {
				t.Fatalf("expected code %d, got %d", tt.status, response.Code)
			}

			var log loggerMsg
			if err := json.NewDecoder(&logBuffer).Decode(&log); err != nil {
				t.Fatalf("error decoding log %v", err)
			}

			t.Logf("logged: %+v", log)
			if log.Msg != "request completed" {
				t.Fatalf("wrong message: %v", log.Msg)
			}
			if log.Method != method {
				t.Fatalf("expected %v, got %v", method, log.Method)
			}
			if log.Path != path {
				t.Fatalf("expected %v, got %v", path, log.Path)
			}
			if log.Status != tt.status {
				t.Fatalf("expected logged status %d, got %d", tt.status, log.Status)
			}
			if log.ClientIP != ip {
				t.Fatalf("expected ip %v, got %v", ip, log.ClientIP)
			}
		})
	}
}
