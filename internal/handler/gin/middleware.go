package gin_handler

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linsfelipec/url-analytics/internal/handler"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()

		handler.LogRequest(
			logger,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(startTime),
			c.Request.RemoteAddr,
		)
	}
}
