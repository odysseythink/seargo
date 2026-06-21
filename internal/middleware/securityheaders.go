package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders returns Gin middleware that sets default security headers,
// but does NOT override headers already set by the handler.
func SecurityHeaders(headers map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for key, value := range headers {
			// Only set if not already written
			if c.Writer.Header().Get(key) == "" {
				c.Header(key, value)
			}
		}
		c.Next()
	}
}
