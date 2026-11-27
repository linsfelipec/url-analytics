package handler

import (
	"log/slog"
	"time"
)

func LogRequest(
	logger *slog.Logger,
	method string,
	path string,
	status int,
	duration time.Duration,
	clientAddr string,
) {
	logger.Info("request completed",
		"method", method,
		"path", path,
		"status", status,
		"duration", float64(duration.Microseconds())/1000,
		"client_addr", clientAddr,
	)
}
