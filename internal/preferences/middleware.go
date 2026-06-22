package preferences

import (
	"github.com/gin-gonic/gin"
)

const ctxKeyPreferences = "seargo.preferences"

// PreferencesMiddleware loads preferences from the cookie and attaches
// them to the Gin context. Cookie decode errors return HTTP 400.
func PreferencesMiddleware(store *PreferencesStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		prefs, err := store.Load(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "invalid_preferences_cookie"})
			return
		}
		c.Set(ctxKeyPreferences, prefs)
		c.Next()
	}
}

// CtxPreferences retrieves the resolved UserPreferences from the Gin context.
// Returns nil if the middleware was not installed.
func CtxPreferences(c *gin.Context) *UserPreferences {
	v, ok := c.Get(ctxKeyPreferences)
	if !ok {
		return nil
	}
	prefs, _ := v.(*UserPreferences)
	return prefs
}
