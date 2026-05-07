package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/metrics"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := strconv.Itoa(c.Writer.Status())
		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration.Seconds())

		logger.Info("http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
