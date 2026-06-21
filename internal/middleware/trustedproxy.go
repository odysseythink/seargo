package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/seargo/seargo/internal/security"
)

// TrustedProxy returns Gin middleware that extracts the real client IP into gin.Context "clientIP".
func TrustedProxy(extractor security.IPExtractor) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := extractor.ClientIP(c.Request)
		c.Set("clientIP", ip)
		c.Next()
	}
}

// HandleRobotsTxt serves the /robots.txt endpoint.
func HandleRobotsTxt(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	c.String(200, `User-agent: *
Disallow: /image_proxy
Disallow: /favicon_proxy
Disallow: /metrics
Disallow: /health
`)
}
