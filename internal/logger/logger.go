package logger

import (
	"io"
	"log/slog"
)

func New(output io.Writer) *slog.Logger {
	opt := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	return slog.New(slog.NewJSONHandler(output, opt))
}
