package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(l *slog.Logger) gin.HandlerFunc {
	if l == nil {
		l = slog.Default()
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		if raw != "" {
			path = path + "?" + raw
		}

		rid, _ := c.Get("request_id")

		l.Info("http request",
			"status", status,
			"method", method,
			"path", path,
			"ip", clientIP,
			"latency_ms", latency.Milliseconds(),
			"request_id", rid,
		)
	}
}
