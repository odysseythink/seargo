package middleware

import (
	"crypto/rand"
	"fmt"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

const maxRequestIDLen = 64
const contextKeyRequestID = "request_id"

// RequestID is a Gin middleware that ensures every request has a unique ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		id = strings.TrimSpace(id)

		if id == "" || !isValidRequestIDChars(id) {
			id = generateUUID()
		} else if len(id) > maxRequestIDLen {
			id = id[:maxRequestIDLen]
		}

		c.Set(contextKeyRequestID, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func isValidRequestIDChars(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' {
			return false
		}
	}
	return true
}

func generateUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
