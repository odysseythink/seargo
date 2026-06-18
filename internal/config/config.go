package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// -------- Top-level Config --------

type Config struct {
	General            GeneralConfig                `yaml:"general"`
	Brand              BrandConfig                  `yaml:"brand"`
	Search             SearchConfig                 `yaml:"search"`
	Server             ServerConfig                 `yaml:"server"`
	Outgoing           OutgoingConfig               `yaml:"outgoing"`
	UI                 UIConfig                     `yaml:"ui"`
	Preferences        PreferencesConfig            `yaml:"preferences"`
	Valkey             ValkeyConfig                 `yaml:"valkey"`
	Plugins            map[string]PluginConfig      `yaml:"plugins"`
	CategoriesAsTabs   map[string]CategoryTabConfig `yaml:"categories_as_tabs"`
	Engines            []EngineConfig               `yaml:"engines"`
	DOIRsolvers        map[string]string            `yaml:"doi_resolvers"`
	DefaultDOIResolver string                       `yaml:"default_doi_resolver"`
	UseDefaultSettings *UseDefaultSettings          `yaml:"use_default_settings"`
	Cache              CacheConfig                  `yaml:"cache"`
}

// -------- Blocks --------

type GeneralConfig struct {
	Debug            bool    `yaml:"debug"`
	InstanceName     string  `yaml:"instance_name"`
	PrivacyPolicyURL *string `yaml:"privacypolicy_url"`
	ContactURL       *string `yaml:"contact_url"`
	DonationURL      string  `yaml:"donation_url"`
	EnableMetrics    bool    `yaml:"enable_metrics"`
	OpenMetrics      string  `yaml:"open_metrics"`
}

type BrandConfig struct {
	IssueURL        string      `yaml:"issue_url"`
	DocsURL         string      `yaml:"docs_url"`
	PublicInstances string      `yaml:"public_instances"`
	WikiURL         string      `yaml:"wiki_url"`
	NewIssueURL     string      `yaml:"new_issue_url"`
	Custom          BrandCustom `yaml:"custom"`
	PWAColors       ThemeColors `yaml:"pwa_colors"`
}

type BrandCustom struct {
	Links map[string]string `yaml:"links"`
}

type ThemeColors struct {
	ThemeColorLight      string `yaml:"theme_color_light"`
	BackgroundColorLight string `yaml:"background_color_light"`
	ThemeColorDark       string `yaml:"theme_color_dark"`
	BackgroundColorDark  string `yaml:"background_color_dark"`
	ThemeColorBlack      string `yaml:"theme_color_black"`
	BackgroundColorBlack string `yaml:"background_color_black"`
}

type SearchConfig struct {
	SafeSearch         int                  `yaml:"safe_search"`
	Autocomplete       string               `yaml:"autocomplete"`
	AutocompleteMin    int                  `yaml:"autocomplete_min"`
	FaviconResolver    string               `yaml:"favicon_resolver"`
	DefaultLang        string               `yaml:"default_lang"`
	Languages          []string             `yaml:"languages"`
	DefaultCategory    string               `yaml:"default_category"`
	MaxResults         int                  `yaml:"max_results"`
	BanTimeOnFail      float64              `yaml:"ban_time_on_fail"`
	MaxBanTimeOnFail   float64              `yaml:"max_ban_time_on_fail"`
	Formats            []string             `yaml:"formats"`
	MaxPage            int                  `yaml:"max_page"`
	SuspendedTimes     SuspendedTimesConfig `yaml:"suspended_times"`
}

type SuspendedTimesConfig struct {
	SearxEngineAccessDenied     float64 `yaml:"SearxEngineAccessDenied"`
	SearxEngineCaptcha          float64 `yaml:"SearxEngineCaptcha"`
	SearxEngineTooManyRequests  float64 `yaml:"SearxEngineTooManyRequests"`
	CfSearxEngineCaptcha        float64 `yaml:"cf_SearxEngineCaptcha"`
	CfSearxEngineAccessDenied   float64 `yaml:"cf_SearxEngineAccessDenied"`
	RecaptchaSearxEngineCaptcha float64 `yaml:"recaptcha_SearxEngineCaptcha"`
}

type ServerConfig struct {
	Port                int               `yaml:"port"`
	BindAddress         string            `yaml:"bind_address"`
	Limiter             bool              `yaml:"limiter"`
	PublicInstance      bool              `yaml:"public_instance"`
	SecretKey           string            `yaml:"secret_key"`
	BaseURL             *string           `yaml:"base_url"`
	ImageProxy          bool              `yaml:"image_proxy"`
	HTTPProtocolVersion string            `yaml:"http_protocol_version"`
	Method              string            `yaml:"method"`
	DefaultHTTPHeaders  map[string]string `yaml:"default_http_headers"`
}

type OutgoingConfig struct {
	UserAgentSuffix   string      `yaml:"useragent_suffix"`
	RequestTimeout    float64     `yaml:"request_timeout"`
	EnableHTTP2       bool        `yaml:"enable_http2"`
	Verify            interface{} `yaml:"verify"`
	MaxRequestTimeout *float64    `yaml:"max_request_timeout"`
	PoolConnections   int         `yaml:"pool_connections"`
	PoolMaxsize       int         `yaml:"pool_maxsize"`
	KeepaliveExpiry   float64     `yaml:"keepalive_expiry"`
	MaxRedirects      int         `yaml:"max_redirects"`
	Retries           int         `yaml:"retries"`
	Proxies           interface{} `yaml:"proxies"`
	SourceIPs         interface{} `yaml:"source_ips"`
	UsingTorProxy     bool        `yaml:"using_tor_proxy"`
	ExtraProxyTimeout int         `yaml:"extra_proxy_timeout"`
	UserAgent         string      `yaml:"useragent"`
	Timeout           int         `yaml:"timeout"`
}

type UIConfig struct {
	StaticPath             string      `yaml:"static_path"`
	TemplatesPath          string      `yaml:"templates_path"`
	DefaultTheme           string      `yaml:"default_theme"`
	DefaultLocale          string      `yaml:"default_locale"`
	CenterAlignment        bool        `yaml:"center_alignment"`
	ResultsOnNewTab        bool        `yaml:"results_on_new_tab"`
	QueryInTitle           bool        `yaml:"query_in_title"`
	CacheURL               string      `yaml:"cache_url"`
	SearchOnCategorySelect bool        `yaml:"search_on_category_select"`
	Hotkeys                string      `yaml:"hotkeys"`
	URLFormatting          string      `yaml:"url_formatting"`
	ThemeArgs              UIThemeArgs `yaml:"theme_args"`
}

type UIThemeArgs struct {
	SimpleStyle string `yaml:"simple_style"`
}

type PreferencesConfig struct {
	Lock []string `yaml:"lock"`
}

type ValkeyConfig struct {
	URL *string `yaml:"url"`
}

type PluginConfig struct {
	Active bool                   `yaml:"active"`
	Extra  map[string]interface{} `yaml:",inline"`
}

type CategoryTabConfig struct {
	Engines []string `yaml:"engines"`
}

type EngineConfig struct {
	Name       string                 `yaml:"name"`
	Engine     string                 `yaml:"engine"`
	Disabled   bool                   `yaml:"disabled"`
	Shortcut   string                 `yaml:"shortcut"`
	Categories []string               `yaml:"categories"`
	Weight     float64                `yaml:"weight"`
	Timeout    float64                `yaml:"timeout"`
	APIKey     string                 `yaml:"api_key"`
	Extra      map[string]interface{} `yaml:"extra"`
	Enabled    bool                   `yaml:"enabled"`
}

type UseDefaultSettings struct {
	Engines UseDefaultSettingsEngines `yaml:"engines"`
}

type UseDefaultSettingsEngines struct {
	Remove   []string `yaml:"remove"`
	KeepOnly []string `yaml:"keep_only"`
}

type CacheConfig struct {
	Enabled   bool   `yaml:"enabled"`
	LocalTTL  int    `yaml:"local_ttl"`
	RedisTTL  int    `yaml:"redis_ttl"`
	RedisAddr string `yaml:"redis_addr"`
}

// -------- Load --------

func Load(path string) (*Config, error) {
	cfg := builtInDefaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var userCfg Config
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	overlayDefaults(&cfg, &userCfg)
	applyEnvOverrides(&cfg)

	if cfg.UseDefaultSettings != nil {
		applyUseDefaultSettings(&cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// -------- Placeholder (filled in Task 2) --------

func overlayDefaults(dst *Config, src *Config) {
	// Stub — full implementation in Task 2
}

// -------- Placeholder (filled in Task 3) --------

func applyUseDefaultSettings(cfg *Config) {
	// Stub — full implementation in Task 3
}

// -------- Validate (minimal for now, expanded in Task 4) --------

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port)
	}
	if c.Search.MaxResults <= 0 {
		c.Search.MaxResults = 10
	}
	if c.Search.SafeSearch < 0 || c.Search.SafeSearch > 2 {
		return fmt.Errorf("search.safe_search must be 0, 1, or 2, got %d", c.Search.SafeSearch)
	}
	return nil
}

// -------- Env overrides (extended in Task 2) --------

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SEARGO_SERVER_SECRET_KEY"); v != "" {
		cfg.Server.SecretKey = v
	}
	if v := os.Getenv("SEARGO_CACHE_REDIS_ADDR"); v != "" {
		cfg.Cache.RedisAddr = v
	}
	if v := os.Getenv("SEARGO_REQUEST_TIMEOUT"); v != "" {
		if t, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Outgoing.RequestTimeout = t
		}
	}
	for i := range cfg.Engines {
		envKey := fmt.Sprintf("SEARGO_ENGINE_%s_API_KEY", strings.ToUpper(cfg.Engines[i].Name))
		if v := os.Getenv(envKey); v != "" {
			cfg.Engines[i].APIKey = v
		}
	}
}
