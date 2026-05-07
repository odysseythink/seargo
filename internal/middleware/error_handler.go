package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/seargo/seargo/internal/errors"
	"github.com/seargo/seargo/internal/logger"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		lastErr := c.Errors.Last().Err
		if appErr, ok := lastErr.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": appErr})
			return
		}

		logger.Error("unhandled error", "error", lastErr, "path", c.Request.URL.Path)
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternal})
	}
}
