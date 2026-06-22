package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/odysseythink/mlog"

	apperrors "github.com/seargo/seargo/internal/errors"
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

		mlog.Error("unhandled error", "error", lastErr, "path", c.Request.URL.Path)
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternal})
	}
}
