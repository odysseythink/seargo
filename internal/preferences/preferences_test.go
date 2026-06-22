package preferences

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seargo/seargo/internal/config"
)

func defaultTestConfig() *config.Config {
	return &config.Config{
		Search: config.SearchConfig{
			DefaultLang:     "en",
			Languages:       []string{"en", "zh-CN", "fr", "de"},
			DefaultCategory: "general",
			SafeSearch:      1,
			Autocomplete:    "google",
			FaviconResolver: "",
			MaxResults:      10,
		},
		UI: config.UIConfig{
			DefaultTheme:    "simple",
			DefaultLocale:   "en",
			CenterAlignment: false,
		},
		Server: config.ServerConfig{
			Method:     "POST",
			ImageProxy: false,
		},
		Preferences: config.PreferencesConfig{
			Lock: []string{},
		},
		Engines: []config.EngineConfig{
			{Name: "google", Engine: "google", Categories: []string{"general", "images"}, Disabled: false},
			{Name: "bing", Engine: "bing", Categories: []string{"general"}, Disabled: false},
			{Name: "duckduckgo", Engine: "duckduckgo", Categories: []string{"general"}, Disabled: true},
		},
		Plugins:   map[string]config.PluginConfig{},
		Answerers: map[string]config.AnswererConfig{},
	}
}

func TestPreferenceValue(t *testing.T) {
	pv := PreferenceValue{StringValue: "zh-CN"}
	if pv.StringValue != "zh-CN" {
		t.Errorf("StringValue = %q, want %q", pv.StringValue, "zh-CN")
	}
}

func TestValidateLanguage(t *testing.T) {
	cfg := defaultTestConfig()
	choices := map[string][]string{
		"language": cfg.Search.Languages,
	}
	locked := map[string]bool{}

	// Valid languages
	if got := validateField("language", "zh-CN", choices, locked, "en"); got != "zh-CN" {
		t.Errorf("valid language zh-CN rejected: got %q", got)
	}
	if got := validateField("language", "fr", choices, locked, "en"); got != "fr" {
		t.Errorf("valid language fr rejected: got %q", got)
	}

	// Invalid language → default
	if got := validateField("language", "__invalid__", choices, locked, "en"); got != "en" {
		t.Errorf("invalid language should default: got %q", got)
	}

	// Locked field → default
	locked["language"] = true
	if got := validateField("language", "zh-CN", choices, locked, "en"); got != "en" {
		t.Errorf("locked field should return default: got %q", got)
	}
}

func TestValidateSafeSearch(t *testing.T) {
	tests := []struct {
		input  string
		wantOk bool
	}{
		{"0", true},
		{"1", true},
		{"2", true},
		{"3", false},
		{"abc", false},
		{"", true}, // empty → OK, kept as-is
	}
	for _, tt := range tests {
		got := validateSafeSearch(tt.input, 1)
		if tt.input == "0" || tt.input == "1" || tt.input == "2" || tt.input == "" {
			if got != tt.input && tt.input != "" {
				t.Errorf("validateSafeSearch(%q) = %q, want %q (keep valid)", tt.input, got, tt.input)
			}
		} else {
			if got != "1" {
				t.Errorf("validateSafeSearch(%q) = %q, want %q (default)", tt.input, got, "1")
			}
		}
		_ = tt.wantOk
	}
}

func TestValidateChoiceField(t *testing.T) {
	choices := map[string][]string{
		"theme": {"simple", "auto", "dark"},
	}
	locked := map[string]bool{}

	if got := validateField("theme", "simple", choices, locked, "simple"); got != "simple" {
		t.Errorf("valid theme rejected: got %q", got)
	}
	if got := validateField("theme", "nonexistent", choices, locked, "simple"); got != "simple" {
		t.Errorf("invalid theme should default: got %q", got)
	}
}

func TestValidateLanguageCode_wrongFormat(t *testing.T) {
	// language code validation: must match configured languages list
	cfg := defaultTestConfig()
	choices := map[string][]string{
		"language": cfg.Search.Languages,
	}
	locked := map[string]bool{}

	// "xx" is not in languages list → falls back to default
	if got := validateField("language", "xx", choices, locked, "en"); got != "en" {
		t.Errorf("unsupported language code should default: got %q", got)
	}
}

func TestPreferencesMiddleware(t *testing.T) {
	cfg := defaultTestConfig()
	store := NewStore(cfg)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(PreferencesMiddleware(store))
	r.GET("/test", func(c *gin.Context) {
		prefs := CtxPreferences(c)
		c.JSON(200, gin.H{"language": prefs.Language})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	cookieRaw := rawPreferences{"language": "fr"}
	encoded, _ := store.codec.Encode(cookieRaw)
	req.AddCookie(&http.Cookie{Name: "seargo_preferences", Value: encoded})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCtxPreferences_NilWhenNoMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		prefs := CtxPreferences(c)
		if prefs != nil {
			t.Error("expected nil when middleware not installed")
		}
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
}
