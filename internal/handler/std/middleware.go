package std_handler

import (
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/linsfelipec/url-analytics/internal/handler"
)

type Middleware func(http.Handler) http.Handler

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func RequestLogger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
			}
			next.ServeHTTP(rw, r)

			handler.LogRequest(
				logger,
				r.Method,
				r.URL.Path,
				rw.statusCode,
				time.Since(startTime),
				r.RemoteAddr,
			)
		})
	}
}

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, v := range slices.Backward(middlewares) {
		handler = v(handler)
	}
	return handler
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if rw.statusCode != 0 {
		return
	}
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
