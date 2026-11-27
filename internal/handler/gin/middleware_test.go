package gin_handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linsfelipec/url-analytics/internal/logger"
)

type loggerMsg struct {
	Msg        string  `json:"msg"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Status     int     `json:"status"`
	DurationUs float64 `json:"duration_us"`
	ClientIP   string  `json:"client_addr"`
}

func TestMiddleware_RequestLogger(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		handler gin.HandlerFunc
	}{
		{
			name:   "write header",
			status: http.StatusCreated,
			handler: func(c *gin.Context) {
				c.Writer.WriteHeader(http.StatusCreated)
			},
		},
		{
			name:   "write call",
			status: http.StatusOK,
			handler: func(c *gin.Context) {
				c.Writer.Write(nil)
			},
		},
		{
			name:   "write call with explicit status",
			status: http.StatusInternalServerError,
			handler: func(c *gin.Context) {
				c.Writer.WriteHeader(http.StatusInternalServerError)
				c.Writer.Write(nil)
			},
		},
	}

	const (
		path   = "/cmd/api/abc123"
		method = http.MethodGet
		ip     = "127.0.0.1:54321"
	)

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			testLogger := logger.New(&logBuffer)

			request, err := http.NewRequest(method, path, nil)
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}
			request.RemoteAddr = ip

			router := gin.New()
			router.Use(gin.Recovery(), RequestLogger(testLogger))
			router.GET("/cmd/api/:code", tt.handler)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

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
