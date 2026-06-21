package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seargo/seargo/internal/botdetection"
)

// BotDetection returns Gin middleware that runs bot detection probes.
func BotDetection(det *botdetection.Detector) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP, _ := c.Get("clientIP")
		ip, _ := clientIP.(string)

		dec, reason, err := det.Filter(c.Request.Context(), c.Request, ip)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "bot detection error"})
			return
		}

		switch dec {
		case botdetection.Block:
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":  "request blocked",
				"reason": reason,
			})
			return
		case botdetection.Redirect:
			c.Redirect(http.StatusFound, "/link_token")
			c.Abort()
			return
		}

		c.Next()
	}
}
