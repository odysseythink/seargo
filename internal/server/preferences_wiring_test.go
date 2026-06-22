package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/preferences"
	"github.com/seargo/seargo/internal/search"
)

// newTestSearchServer creates a test server with preferences middleware,
// search route, and preferences routes (so we can set cookies via PUT).
func newTestSearchServer(cfg *config.Config, prefsStore *preferences.PreferencesStore) *Server {
	scheduler, err := search.NewScheduler(cfg, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		panic(err)
	}
	s := &Server{
		router:           gin.New(),
		config:           cfg,
		scheduler:        scheduler,
		preferencesStore: prefsStore,
	}
	gin.SetMode(gin.TestMode)
	r := s.router
	r.Use(preferences.PreferencesMiddleware(prefsStore))
	api := r.Group("/api")
	api.GET("/search", s.handleSearch)
	api.GET("/preferences", s.handleGetPreferences)
	api.PUT("/preferences", s.handlePutPreferences)
	return s
}

// getPrefsCookieValue sets preferences via PUT /api/preferences and returns the cookie value.
func getPrefsCookieValue(t *testing.T, s *Server, body string) string {
	t.Helper()
	req := httptest.NewRequest("PUT", "/api/preferences", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT preferences returned %d: %s", w.Code, w.Body.String())
	}

	for _, c := range w.Result().Cookies() {
		if c.Name == "seargo_preferences" {
			return c.Value
		}
	}
	t.Fatal("no seargo_preferences cookie in PUT response")
	return ""
}

func TestHandleSearch_NoCookieDefaultsToConfig(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestSearchServer(cfg, store)

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	// Should return a valid search response, not an error
	if errMsg, ok := resp["error"]; ok {
		t.Errorf("unexpected error in response: %v", errMsg)
	}
}

func TestHandleSearch_CorruptedCookieReturns400(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestSearchServer(cfg, store)

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "seargo_preferences",
		Value: "!!!corrupted!!!",
	})
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for corrupted cookie, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSearch_CookieCategoryOverridesDefault(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestSearchServer(cfg, store)

	// Set category preference via PUT
	// Note: the preferences API expects categories as an array
	cookieValue := getPrefsCookieValue(t, s, `{"categories":["news"]}`)

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "seargo_preferences",
		Value: cookieValue,
	})
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleSearch_NoMiddlewareInstalled — regression: if middleware is not installed,
// handleSearch should still work with config defaults (CtxPreferences returns nil).
func TestHandleSearch_NoMiddlewareInstalled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := minimalTestConfig()
	scheduler, err := search.NewScheduler(cfg, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{
		router:    gin.New(),
		config:    cfg,
		scheduler: scheduler,
	}
	s.router.GET("/api/search", s.handleSearch)

	req := httptest.NewRequest("GET", "/api/search?q=test", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 without middleware, got %d: %s", w.Code, w.Body.String())
	}
}
