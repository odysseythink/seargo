package preferences

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/seargo/seargo/internal/config"
)

const (
	cookieName   = "seargo_preferences"
	cookieMaxAge = 5 * 365 * 24 * 3600 // 5 years in seconds
)

// PreferencesStore holds config defaults, valid choices, and a codec.
type PreferencesStore struct {
	cfg           *config.Config
	codec         CookieCodec
	validChoices  map[string][]string
	defaultValues rawPreferences
	locked        map[string]bool
}

// NewStore creates a PreferencesStore from the server config.
func NewStore(cfg *config.Config) *PreferencesStore {
	store := &PreferencesStore{
		cfg:    cfg,
		locked: make(map[string]bool),
	}

	for _, key := range cfg.Preferences.Lock {
		store.locked[key] = true
	}

	store.validChoices = map[string][]string{
		"language":     cfg.Search.Languages,
		"theme":        {"auto", "light", "dark", "black"},
		"autocomplete": {"google", "bing", "duckduckgo", "brave", "qwant", "startpage",
			"wikipedia", "dbpedia", "swisscows", "baidu", "naver", "yandex",
			"seznam", "sogou", "mwmbl", "privacywall", "360search", "quark"},
		"method":     {"GET", "POST"},
		"safesearch": {"0", "1", "2"},
	}

	store.defaultValues = rawPreferences{
		"language":                 cfg.Search.DefaultLang,
		"locale":                   cfg.UI.DefaultLocale,
		"autocomplete":             cfg.Search.Autocomplete,
		"safesearch":               strconv.Itoa(cfg.Search.SafeSearch),
		"theme":                    cfg.UI.DefaultTheme,
		"method":                   cfg.Server.Method,
		"image_proxy":              boolToString(cfg.Server.ImageProxy),
		"center_alignment":         boolToString(cfg.UI.CenterAlignment),
		"results_on_new_tab":       boolToString(cfg.UI.ResultsOnNewTab),
		"query_in_title":           boolToString(cfg.UI.QueryInTitle),
		"search_on_category_select": boolToString(cfg.UI.SearchOnCategorySelect),
		"hotkeys":                  "default",
		"url_formatting":           cfg.UI.URLFormatting,
		"simple_style":             cfg.UI.ThemeArgs.SimpleStyle,
		"favicon_resolver":         cfg.Search.FaviconResolver,
		"doi_resolver":             "",
	}

	if store.defaultValues["locale"] == "" {
		store.defaultValues["locale"] = store.defaultValues["language"]
	}

	return store
}

// Load reads the cookie from the request, decodes it, validates and merges with defaults.
func (s *PreferencesStore) Load(req *http.Request) (*UserPreferences, error) {
	raw := make(rawPreferences, len(s.defaultValues))
	for k, v := range s.defaultValues {
		raw[k] = v
	}

	cookie, err := req.Cookie(cookieName)
	if err == nil && cookie.Value != "" {
		cookieRaw, decodeErr := s.codec.Decode(cookie.Value)
		if decodeErr != nil {
			return nil, fmt.Errorf("invalid_preferences_cookie: %w", decodeErr)
		}
		// Merge cookie values over defaults
		for k, v := range cookieRaw {
			if v == "" {
				continue
			}
			if def, ok := s.defaultValues[k]; ok {
				validated := validateField(k, v, s.validChoices, s.locked, def)
				raw[k] = validated
			}
		}
	}

	return buildUserPreferences(raw, s.locked), nil
}

// ApplyUpdate merges an update into current preferences, validates, and returns the new state.
func (s *PreferencesStore) ApplyUpdate(current *UserPreferences, update PreferencesUpdate) (*UserPreferences, error) {
	raw := userPrefsToRaw(current)

	if update.Language != nil && !s.locked["language"] {
		raw["language"] = validateField("language", *update.Language, s.validChoices, s.locked, s.defaultValues["language"])
	}
	if update.Locale != nil && !s.locked["locale"] {
		raw["locale"] = *update.Locale
	}
	if update.Theme != nil && !s.locked["theme"] {
		raw["theme"] = validateField("theme", *update.Theme, s.validChoices, s.locked, s.defaultValues["theme"])
	}
	if update.SafeSearch != nil && !s.locked["safesearch"] {
		raw["safesearch"] = validateField("safesearch", strconv.Itoa(*update.SafeSearch), s.validChoices, s.locked, s.defaultValues["safesearch"])
	}
	if update.Autocomplete != nil && !s.locked["autocomplete"] {
		raw["autocomplete"] = validateField("autocomplete", *update.Autocomplete, s.validChoices, s.locked, s.defaultValues["autocomplete"])
	}
	if update.Method != nil && !s.locked["method"] {
		raw["method"] = validateField("method", *update.Method, s.validChoices, s.locked, s.defaultValues["method"])
	}
	if update.ImageProxy != nil {
		raw["image_proxy"] = boolToString(*update.ImageProxy)
	}
	if update.CenterAlignment != nil {
		raw["center_alignment"] = boolToString(*update.CenterAlignment)
	}
	if update.ResultsOnNewTab != nil {
		raw["results_on_new_tab"] = boolToString(*update.ResultsOnNewTab)
	}
	if update.QueryInTitle != nil {
		raw["query_in_title"] = boolToString(*update.QueryInTitle)
	}
	if update.SearchOnCategorySelect != nil {
		raw["search_on_category_select"] = boolToString(*update.SearchOnCategorySelect)
	}
	if update.FaviconResolver != nil {
		raw["favicon_resolver"] = *update.FaviconResolver
	}
	if update.DOIResolver != nil {
		raw["doi_resolver"] = *update.DOIResolver
	}
	if update.SimpleStyle != nil {
		raw["simple_style"] = *update.SimpleStyle
	}
	if update.Hotkeys != nil {
		raw["hotkeys"] = *update.Hotkeys
	}
	if update.URLFormatting != nil {
		raw["url_formatting"] = *update.URLFormatting
	}
	if update.DisabledEngines != nil {
		raw["disabled_engines"] = commaJoin(update.DisabledEngines)
	}
	if update.EnabledEngines != nil {
		raw["enabled_engines"] = commaJoin(update.EnabledEngines)
	}
	if update.DisabledPlugins != nil {
		raw["disabled_plugins"] = commaJoin(update.DisabledPlugins)
	}
	if update.EnabledPlugins != nil {
		raw["enabled_plugins"] = commaJoin(update.EnabledPlugins)
	}
	if update.Tokens != nil {
		raw["tokens"] = commaJoin(update.Tokens)
	}
	if update.Categories != nil {
		for i, c := range update.Categories {
			raw[fmt.Sprintf("category_%d", i)] = c
		}
	}

	return buildUserPreferences(raw, s.locked), nil
}

// WriteCookie encodes the preferences and sets the Set-Cookie header on the response.
func (s *PreferencesStore) WriteCookie(prefs *UserPreferences, w http.ResponseWriter) error {
	raw := userPrefsToRaw(prefs)
	encoded, err := s.codec.Encode(raw)
	if err != nil {
		return fmt.Errorf("encode cookie: %w", err)
	}
	if len(encoded) > 4096 {
		return fmt.Errorf("preferences_too_large: encoded cookie is %d bytes (max 4096)", len(encoded))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		MaxAge:   cookieMaxAge,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
		Secure:   false,
	})
	return nil
}

// ExportURL returns a settings-as-URL string for sharing/importing preferences.
func (s *PreferencesStore) ExportURL(prefs *UserPreferences) (string, error) {
	raw := userPrefsToRaw(prefs)

	values := url.Values{}
	for k, v := range raw {
		values.Set(k, v)
	}

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write([]byte(values.Encode())); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	encoded := base64.RawURLEncoding.EncodeToString(compressed.Bytes())
	return encoded, nil
}

// ImportURL decodes a settings-as-URL string and returns the merged preferences.
func (s *PreferencesStore) ImportURL(blob string) (*UserPreferences, error) {
	r := base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(blob))
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("import zlib: %w", err)
	}
	defer zr.Close()
	rawBytes, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("import read: %w", err)
	}
	values, err := url.ParseQuery(string(rawBytes))
	if err != nil {
		return nil, fmt.Errorf("import parse: %w", err)
	}

	raw := make(rawPreferences, len(s.defaultValues))
	for k, v := range s.defaultValues {
		raw[k] = v
	}
	for k, vs := range values {
		if len(vs) > 0 && vs[0] != "" {
			if def, ok := s.defaultValues[k]; ok {
				raw[k] = validateField(k, vs[0], s.validChoices, s.locked, def)
			}
		}
	}

	return buildUserPreferences(raw, s.locked), nil
}

// --- helper functions ---

func buildUserPreferences(raw rawPreferences, locked map[string]bool) *UserPreferences {
	ss, _ := strconv.Atoi(raw["safesearch"])
	return &UserPreferences{
		Categories:             commaSplit(raw["category"]),
		Language:               raw["language"],
		Locale:                 raw["locale"],
		Autocomplete:           raw["autocomplete"],
		FaviconResolver:        raw["favicon_resolver"],
		ImageProxy:             stringToBool(raw["image_proxy"]),
		Method:                 raw["method"],
		SafeSearch:             ss,
		Theme:                  raw["theme"],
		ResultsOnNewTab:        stringToBool(raw["results_on_new_tab"]),
		DOIResolver:            raw["doi_resolver"],
		SimpleStyle:            raw["simple_style"],
		CenterAlignment:        stringToBool(raw["center_alignment"]),
		QueryInTitle:           stringToBool(raw["query_in_title"]),
		SearchOnCategorySelect: stringToBool(raw["search_on_category_select"]),
		Hotkeys:                raw["hotkeys"],
		URLFormatting:          raw["url_formatting"],
		DisabledEngines:        commaSplit(raw["disabled_engines"]),
		EnabledEngines:         commaSplit(raw["enabled_engines"]),
		DisabledPlugins:        commaSplit(raw["disabled_plugins"]),
		EnabledPlugins:         commaSplit(raw["enabled_plugins"]),
		Tokens:                 commaSplit(raw["tokens"]),
		Locked:                 keysOf(locked),
	}
}

func userPrefsToRaw(prefs *UserPreferences) rawPreferences {
	return rawPreferences{
		"language":                  prefs.Language,
		"locale":                    prefs.Locale,
		"autocomplete":              prefs.Autocomplete,
		"favicon_resolver":          prefs.FaviconResolver,
		"image_proxy":               boolToString(prefs.ImageProxy),
		"method":                    prefs.Method,
		"safesearch":                strconv.Itoa(prefs.SafeSearch),
		"theme":                     prefs.Theme,
		"results_on_new_tab":        boolToString(prefs.ResultsOnNewTab),
		"doi_resolver":              prefs.DOIResolver,
		"simple_style":              prefs.SimpleStyle,
		"center_alignment":          boolToString(prefs.CenterAlignment),
		"query_in_title":            boolToString(prefs.QueryInTitle),
		"search_on_category_select": boolToString(prefs.SearchOnCategorySelect),
		"hotkeys":                   prefs.Hotkeys,
		"url_formatting":            prefs.URLFormatting,
		"disabled_engines":          commaJoin(prefs.DisabledEngines),
		"enabled_engines":           commaJoin(prefs.EnabledEngines),
		"disabled_plugins":          commaJoin(prefs.DisabledPlugins),
		"enabled_plugins":           commaJoin(prefs.EnabledPlugins),
		"tokens":                    commaJoin(prefs.Tokens),
	}
}

// engineEnabled checks if an engine is enabled for a given category.
func engineEnabled(engineName, category string, defaults map[string]bool, disabled, enabled []string) bool {
	key := engineName + "__" + category
	for _, d := range disabled {
		if d == key {
			return false
		}
	}
	for _, e := range enabled {
		if e == key {
			return true
		}
	}
	return defaults[key]
}

func pluginEnabled(pluginID string, defaults map[string]bool, disabled, enabled []string) bool {
	for _, d := range disabled {
		if d == pluginID {
			return false
		}
	}
	for _, e := range enabled {
		if e == pluginID {
			return true
		}
	}
	return defaults[pluginID]
}

func ValidateToken(userTokens, engineTokens []string) bool {
	if len(engineTokens) == 0 {
		return true
	}
	for _, t := range userTokens {
		for _, et := range engineTokens {
			if t == et {
				return true
			}
		}
	}
	return false
}

func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func stringToBool(s string) bool {
	return s == "1" || s == "true"
}

func keysOf(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
