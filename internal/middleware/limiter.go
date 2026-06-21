package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/limiter"
)

// Limiter returns Gin middleware that applies rate limiting.
func Limiter(cfg *config.Config, lm limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP, _ := c.Get("clientIP")
		ip, _ := clientIP.(string)

		isAPI := !isHTMLRequest(c)

		allowed, reason, err := lm.Allow(c.Request.Context(), ip, isAPI)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limiter error"})
			return
		}
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(time.Minute.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":  "rate limit exceeded",
				"reason": reason,
			})
			return
		}

		c.Next()
	}
}

// HandleLimiterLinkToken handles the /link_token endpoint.
func HandleLimiterLinkToken(c *gin.Context) {
	// If the limiter service is stored in the context, use it.
	svc, exists := c.Get("limiterSvc")
	if !exists {
		c.JSON(200, gin.H{"token": "disabled"})
		return
	}
	lm, ok := svc.(limiter.Limiter)
	if !ok {
		c.JSON(200, gin.H{"token": "disabled"})
		return
	}
	token, err := lm.Token(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": token})
}

// isHTMLRequest checks if the request expects an HTML response.
func isHTMLRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html")
}
