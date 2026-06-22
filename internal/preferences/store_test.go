package preferences

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStoreLoad_DefaultOnly(t *testing.T) {
	cfg := defaultTestConfig()
	store := NewStore(cfg)
	req := httptest.NewRequest("GET", "/", nil)

	prefs, err := store.Load(req)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if prefs.Language != "en" {
		t.Errorf("default language = %q, want %q", prefs.Language, "en")
	}
	if prefs.Locale != "en" {
		t.Errorf("default locale = %q, want %q", prefs.Locale, "en")
	}
	if prefs.Theme != "simple" {
		t.Errorf("default theme = %q, want %q", prefs.Theme, "simple")
	}
}

func TestStoreLoad_CookieOverride(t *testing.T) {
	cfg := defaultTestConfig()
	store := NewStore(cfg)

	cookieRaw := rawPreferences{"language": "zh-CN", "locale": "zh-Hans-CN"}
	encoded, _ := store.codec.Encode(cookieRaw)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "seargo_preferences",
		Value: encoded,
	})

	prefs, err := store.Load(req)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if prefs.Language != "zh-CN" {
		t.Errorf("cookie language = %q, want %q", prefs.Language, "zh-CN")
	}
	if prefs.Locale != "zh-Hans-CN" {
		t.Errorf("cookie locale = %q, want %q", prefs.Locale, "zh-Hans-CN")
	}
	if prefs.Theme != "simple" {
		t.Errorf("theme = %q, want %q", prefs.Theme, "simple")
	}
}

func TestStoreLoad_LockedField(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Preferences.Lock = []string{"language"}
	store := NewStore(cfg)

	cookieRaw := rawPreferences{"language": "zh-CN"}
	encoded, _ := store.codec.Encode(cookieRaw)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "seargo_preferences",
		Value: encoded,
	})

	prefs, err := store.Load(req)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if prefs.Language != "en" {
		t.Errorf("locked language = %q, want config default %q", prefs.Language, "en")
	}
}

func TestStoreLoad_CorruptedCookie(t *testing.T) {
	cfg := defaultTestConfig()
	store := NewStore(cfg)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "seargo_preferences",
		Value: "!!!not-valid-base64!!!",
	})

	_, err := store.Load(req)
	if err == nil {
		t.Error("expected error for corrupted cookie")
	}
}

func TestStoreApplyUpdate_PartialUpdate(t *testing.T) {
	cfg := defaultTestConfig()
	store := NewStore(cfg)

	current := &UserPreferences{Language: "en", Locale: "en", Theme: "simple", SafeSearch: 1}

	langZH := "zh-CN"
	update := PreferencesUpdate{Language: &langZH}

	next, err := store.ApplyUpdate(current, update)
	if err != nil {
		t.Fatalf("ApplyUpdate failed: %v", err)
	}
	if next.Language != "zh-CN" {
		t.Errorf("Language = %q, want %q", next.Language, "zh-CN")
	}
	if next.Locale != "en" {
		t.Errorf("Locale = %q, want %q (unchanged)", next.Locale, "en")
	}
	if next.Theme != "simple" {
		t.Errorf("Theme = %q, want %q (unchanged)", next.Theme, "simple")
	}
}

func TestStoreApplyUpdate_LockedField(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Preferences.Lock = []string{"language"}
	store := NewStore(cfg)

	current := &UserPreferences{Language: "en", Locale: "en"}
	langZH := "zh-CN"
	update := PreferencesUpdate{Language: &langZH}

	next, err := store.ApplyUpdate(current, update)
	if err != nil {
		t.Fatalf("ApplyUpdate failed: %v", err)
	}
	if next.Language != "en" {
		t.Errorf("locked Language = %q, want %q", next.Language, "en")
	}
}

func TestStoreWriteCookie(t *testing.T) {
	cfg := defaultTestConfig()
	store := NewStore(cfg)

	prefs := &UserPreferences{Language: "zh-CN", Locale: "zh-Hans-CN", Theme: "simple", SafeSearch: 1}
	w := httptest.NewRecorder()

	err := store.WriteCookie(prefs, w)
	if err != nil {
		t.Fatalf("WriteCookie failed: %v", err)
	}

	cookies := w.Result().Cookies()
	var prefCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "seargo_preferences" {
			prefCookie = c
			break
		}
	}
	if prefCookie == nil {
		t.Fatal("no seargo_preferences cookie set")
	}
	if prefCookie.MaxAge != int((5 * 365 * 24 * time.Hour).Seconds()) {
		t.Errorf("MaxAge = %d, want ~5 years", prefCookie.MaxAge)
	}
	if prefCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", prefCookie.SameSite)
	}
	if prefCookie.HttpOnly {
		t.Error("HttpOnly should be false")
	}
	if prefCookie.Secure {
		t.Error("Secure should be false")
	}
}

func TestStoreExportImportURL(t *testing.T) {
	cfg := defaultTestConfig()
	store := NewStore(cfg)

	prefs := &UserPreferences{
		Language:   "zh-CN",
		Locale:     "zh-Hans-CN",
		SafeSearch: 1,
		Theme:      "simple",
	}

	url, err := store.ExportURL(prefs)
	if err != nil {
		t.Fatalf("ExportURL failed: %v", err)
	}
	if url == "" {
		t.Fatal("ExportURL returned empty string")
	}

	imported, err := store.ImportURL(url)
	if err != nil {
		t.Fatalf("ImportURL failed: %v", err)
	}
	if imported.Language != "zh-CN" {
		t.Errorf("imported Language = %q, want %q", imported.Language, "zh-CN")
	}
	if imported.Locale != "zh-Hans-CN" {
		t.Errorf("imported Locale = %q, want %q", imported.Locale, "zh-Hans-CN")
	}
}

func TestEngineEnabled(t *testing.T) {
	got := engineEnabled("google", "general",
		map[string]bool{"google__general": true},
		[]string{"google__general"}, nil)
	if got {
		t.Error("google__general should be disabled when in disabled_engines")
	}

	got = engineEnabled("duckduckgo", "general",
		map[string]bool{"duckduckgo__general": false},
		nil, []string{"duckduckgo__general"})
	if !got {
		t.Error("duckduckgo__general should be enabled when in enabled_engines")
	}
}

func TestCookieSize(t *testing.T) {
	cfg := defaultTestConfig()
	store := NewStore(cfg)

	prefs := &UserPreferences{
		Language:         "zh-CN",
		Locale:           "zh-Hans-CN",
		SafeSearch:       1,
		Theme:            "simple",
		Autocomplete:     "google",
		DisabledEngines:  []string{"bing__general", "wikipedia__general"},
		EnabledEngines:   []string{"google__general", "google__images", "google__news", "duckduckgo__general", "duckduckgo__images", "brave__general", "brave__images", "yahoo__general"},
		DisabledPlugins:  []string{},
		EnabledPlugins:   []string{"oa_doi_rewrite", "self_info", "url_unshorten", "tor_check_plugin", "search_on_category_select", "hostname_replace", "basic_calculator", "unit_converter", "tracker_url_remover"},
		Tokens:           []string{"token-a", "token-b"},
	}

	raw := userPrefsToRaw(prefs)
	encoded, _ := store.codec.Encode(raw)
	if len(encoded) > 4096 {
		t.Errorf("cookie size %d exceeds 4096 byte limit", len(encoded))
	}
	t.Logf("maximal cookie size: %d bytes", len(encoded))
}
