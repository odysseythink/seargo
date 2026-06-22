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

func newTestServerWithPrefs(cfg *config.Config, prefsStore *preferences.PreferencesStore) *Server {
	scheduler, err := search.NewScheduler(cfg, nil, nil, nil, nil, nil, nil)
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
	api.GET("/preferences", s.handleGetPreferences)
	api.PUT("/preferences", s.handlePutPreferences)
	return s
}

func minimalTestConfig() *config.Config {
	return &config.Config{
		Search: config.SearchConfig{
			DefaultLang:     "en",
			Languages:       []string{"en", "zh-CN"},
			DefaultCategory: "general",
			SafeSearch:      1,
			Autocomplete:    "google",
			MaxResults:      10,
		},
		UI: config.UIConfig{
			DefaultTheme:  "simple",
			DefaultLocale: "en",
		},
		Server: config.ServerConfig{
			Method: "POST",
		},
		Preferences: config.PreferencesConfig{
			Lock: []string{},
		},
		Engines:   []config.EngineConfig{},
		Plugins:   map[string]config.PluginConfig{},
		Answerers: map[string]config.AnswererConfig{},
	}
}

func TestHandleGetPreferences_Default(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestServerWithPrefs(cfg, store)

	req := httptest.NewRequest("GET", "/api/preferences", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["settings"]; !ok {
		t.Error("response missing 'settings' field")
	}
}

func TestHandleGetPreferences_CorruptedCookie(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestServerWithPrefs(cfg, store)

	req := httptest.NewRequest("GET", "/api/preferences", nil)
	req.AddCookie(&http.Cookie{
		Name:  "seargo_preferences",
		Value: "!!!corrupted!!!",
	})
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for corrupted cookie, got %d", w.Code)
	}
}

func TestHandlePutPreferences_ValidUpdate(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestServerWithPrefs(cfg, store)

	body := `{"language":"zh-CN"}`
	req := httptest.NewRequest("PUT", "/api/preferences", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "seargo_preferences" {
			found = true
			break
		}
	}
	if !found {
		t.Error("PUT should set seargo_preferences cookie")
	}
}

func TestHandleExportPreferences(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestServerWithPrefs(cfg, store)
	s.router.GET("/api/preferences/export", s.handleExportPreferences)

	req := httptest.NewRequest("GET", "/api/preferences/export", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("export returned empty body")
	}
}

func TestHandleImportPreferences(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestServerWithPrefs(cfg, store)
	s.router.GET("/api/preferences/export", s.handleExportPreferences)
	s.router.GET("/api/preferences/import", s.handleImportPreferences)

	// First export to get a blob
	req := httptest.NewRequest("GET", "/api/preferences/export", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("export: expected 200, got %d", w.Code)
	}
	blob := w.Body.String()

	// Then import it
	req2 := httptest.NewRequest("GET", "/api/preferences/import?blob="+blob, nil)
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Fatalf("import: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	// Verify a cookie was set
	cookies := w2.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "seargo_preferences" {
			found = true
			break
		}
	}
	if !found {
		t.Error("import should set seargo_preferences cookie")
	}
}

func TestHandleImportPreferences_MissingBlob(t *testing.T) {
	cfg := minimalTestConfig()
	store := preferences.NewStore(cfg)
	s := newTestServerWithPrefs(cfg, store)
	s.router.GET("/api/preferences/import", s.handleImportPreferences)

	req := httptest.NewRequest("GET", "/api/preferences/import", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandlePutPreferences_LockedField(t *testing.T) {
	cfg := minimalTestConfig()
	cfg.Preferences.Lock = []string{"language"}
	store := preferences.NewStore(cfg)
	s := newTestServerWithPrefs(cfg, store)

	body := `{"language":"zh-CN"}`
	req := httptest.NewRequest("PUT", "/api/preferences", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	settings := resp["settings"].(map[string]interface{})
	if settings["language"] != "en" {
		t.Errorf("locked language should remain 'en', got %v", settings["language"])
	}
}
