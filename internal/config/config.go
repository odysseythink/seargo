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

// OutgoingNetworkOverride 用于 outgoing.networks.<name> 配置覆盖。
type OutgoingNetworkOverride struct {
	EnableHTTP              *bool       `yaml:"enable_http"`
	Verify                  *bool       `yaml:"verify"`
	EnableHTTP2             *bool       `yaml:"enable_http2"`
	MaxConnections          *int        `yaml:"max_connections"`
	MaxKeepaliveConnections *int        `yaml:"max_keepalive_connections"`
	KeepaliveExpiry         *float64    `yaml:"keepalive_expiry"`
	LocalAddresses          interface{} `yaml:"local_addresses"`
	Proxies                 interface{} `yaml:"proxies"`
	UsingTorProxy           *bool       `yaml:"using_tor_proxy"`
	MaxRedirects            *int        `yaml:"max_redirects"`
	Retries                 *int        `yaml:"retries"`
	RetryOnHTTPError        interface{} `yaml:"retry_on_http_error"`
	UserAgent               string      `yaml:"useragent"`
	RequestTimeout          *float64    `yaml:"request_timeout"`
	Timeout                 *float64    `yaml:"timeout"`
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
	UserAgent         string                            `yaml:"useragent"`
	Timeout           int                               `yaml:"timeout"`
	// Phase 3 — Network Layer
	EnableHTTP       bool                             `yaml:"enable_http"`          // 是否允许 HTTP；默认 true
	Networks         map[string]OutgoingNetworkOverride `yaml:"networks"`            // 自定义网络
	RetryOnHTTPError interface{}                      `yaml:"retry_on_http_error"`  // nil | bool | int | []int
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

	// SearXNG-compatible fields
	Paging               bool        `yaml:"paging"`
	TimeRangeSupport     bool        `yaml:"time_range_support"`
	LanguageSupport      bool        `yaml:"language_support"`
	SafeSearch           bool        `yaml:"safesearch"`
	DisplayErrorMessages bool        `yaml:"display_error_messages"`
	EnableHTTP           bool        `yaml:"enable_http"`
	Inactive             bool        `yaml:"inactive"`
	Tokens               []string    `yaml:"tokens"`
	Network              string      `yaml:"network"`
	ShortCut             string      `yaml:"short_cut"`
	SoftMaxRedirects     int         `yaml:"soft_max_redirects"`
	NoResultForHTTPStatus []int      `yaml:"no_result_for_http_status"`
	RaiseForHTTPError    interface{} `yaml:"raise_for_http_error"`
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

var validCategories = map[string]bool{
	"general": true, "images": true, "videos": true, "news": true,
	"map": true, "music": true, "it": true, "science": true,
	"files": true, "social media": true,
}

var validHTTPVersions = map[string]bool{"1.0": true, "1.1": true}
var validMethods = map[string]bool{"GET": true, "POST": true}

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

// -------- overlayDefaults --------

func overlayDefaults(dst *Config, src *Config) {
	// General
	if src.General.InstanceName != "" {
		dst.General.InstanceName = src.General.InstanceName
	}
	if src.General.Debug {
		dst.General.Debug = true
	}
	if src.General.PrivacyPolicyURL != nil {
		dst.General.PrivacyPolicyURL = src.General.PrivacyPolicyURL
	}
	if src.General.ContactURL != nil {
		dst.General.ContactURL = src.General.ContactURL
	}
	if src.General.DonationURL != "" {
		dst.General.DonationURL = src.General.DonationURL
	}
	if src.General.EnableMetrics {
		dst.General.EnableMetrics = true
	}
	if src.General.OpenMetrics != "" {
		dst.General.OpenMetrics = src.General.OpenMetrics
	}

	// Brand
	overlayBrand(&dst.Brand, &src.Brand)

	// Search
	overlaySearch(&dst.Search, &src.Search)

	// Server
	overlayServer(&dst.Server, &src.Server)

	// Outgoing
	overlayOutgoing(&dst.Outgoing, &src.Outgoing)

	// UI
	overlayUI(&dst.UI, &src.UI)

	// Preferences
	if len(src.Preferences.Lock) > 0 {
		dst.Preferences.Lock = src.Preferences.Lock
	}

	// Valkey
	if src.Valkey.URL != nil {
		dst.Valkey.URL = src.Valkey.URL
	}

	// Plugins — merge maps
	if src.Plugins != nil {
		if dst.Plugins == nil {
			dst.Plugins = make(map[string]PluginConfig)
		}
		for k, v := range src.Plugins {
			dst.Plugins[k] = v
		}
	}

	// CategoriesAsTabs — merge maps
	if src.CategoriesAsTabs != nil {
		if dst.CategoriesAsTabs == nil {
			dst.CategoriesAsTabs = make(map[string]CategoryTabConfig)
		}
		for k, v := range src.CategoriesAsTabs {
			dst.CategoriesAsTabs[k] = v
		}
	}

	// Engines — replace list if user provided any
	if len(src.Engines) > 0 {
		dst.Engines = src.Engines
	}

	// DOIRsolvers — merge maps
	if src.DOIRsolvers != nil {
		if dst.DOIRsolvers == nil {
			dst.DOIRsolvers = make(map[string]string)
		}
		for k, v := range src.DOIRsolvers {
			dst.DOIRsolvers[k] = v
		}
	}

	// DefaultDOIResolver
	if src.DefaultDOIResolver != "" {
		dst.DefaultDOIResolver = src.DefaultDOIResolver
	}

	// UseDefaultSettings — copy pointer if set
	if src.UseDefaultSettings != nil {
		dst.UseDefaultSettings = src.UseDefaultSettings
	}

	// Legacy: Cache
	if src.Cache.Enabled {
		dst.Cache.Enabled = true
	}
	if src.Cache.LocalTTL > 0 {
		dst.Cache.LocalTTL = src.Cache.LocalTTL
	}
	if src.Cache.RedisTTL > 0 {
		dst.Cache.RedisTTL = src.Cache.RedisTTL
	}
	if src.Cache.RedisAddr != "" {
		dst.Cache.RedisAddr = src.Cache.RedisAddr
	}
}

func overlayBrand(dst *BrandConfig, src *BrandConfig) {
	if src.IssueURL != "" {
		dst.IssueURL = src.IssueURL
	}
	if src.DocsURL != "" {
		dst.DocsURL = src.DocsURL
	}
	if src.PublicInstances != "" {
		dst.PublicInstances = src.PublicInstances
	}
	if src.WikiURL != "" {
		dst.WikiURL = src.WikiURL
	}
	if src.NewIssueURL != "" {
		dst.NewIssueURL = src.NewIssueURL
	}
	if len(src.Custom.Links) > 0 {
		dst.Custom.Links = src.Custom.Links
	}
	if src.PWAColors.ThemeColorLight != "" {
		dst.PWAColors.ThemeColorLight = src.PWAColors.ThemeColorLight
	}
	if src.PWAColors.BackgroundColorLight != "" {
		dst.PWAColors.BackgroundColorLight = src.PWAColors.BackgroundColorLight
	}
	if src.PWAColors.ThemeColorDark != "" {
		dst.PWAColors.ThemeColorDark = src.PWAColors.ThemeColorDark
	}
	if src.PWAColors.BackgroundColorDark != "" {
		dst.PWAColors.BackgroundColorDark = src.PWAColors.BackgroundColorDark
	}
	if src.PWAColors.ThemeColorBlack != "" {
		dst.PWAColors.ThemeColorBlack = src.PWAColors.ThemeColorBlack
	}
	if src.PWAColors.BackgroundColorBlack != "" {
		dst.PWAColors.BackgroundColorBlack = src.PWAColors.BackgroundColorBlack
	}
}

func overlaySearch(dst *SearchConfig, src *SearchConfig) {
	if src.SafeSearch != 0 {
		dst.SafeSearch = src.SafeSearch
	}
	if src.Autocomplete != "" {
		dst.Autocomplete = src.Autocomplete
	}
	if src.AutocompleteMin != 0 {
		dst.AutocompleteMin = src.AutocompleteMin
	}
	if src.FaviconResolver != "" {
		dst.FaviconResolver = src.FaviconResolver
	}
	if src.DefaultLang != "" {
		dst.DefaultLang = src.DefaultLang
	}
	if len(src.Languages) > 0 {
		dst.Languages = src.Languages
	}
	if src.DefaultCategory != "" {
		dst.DefaultCategory = src.DefaultCategory
	}
	if src.MaxResults != 0 {
		dst.MaxResults = src.MaxResults
	}
	if src.BanTimeOnFail != 0 {
		dst.BanTimeOnFail = src.BanTimeOnFail
	}
	if src.MaxBanTimeOnFail != 0 {
		dst.MaxBanTimeOnFail = src.MaxBanTimeOnFail
	}
	if len(src.Formats) > 0 {
		dst.Formats = src.Formats
	}
	if src.MaxPage != 0 {
		dst.MaxPage = src.MaxPage
	}
	overlaySuspendedTimes(&dst.SuspendedTimes, &src.SuspendedTimes)
}

func overlaySuspendedTimes(dst *SuspendedTimesConfig, src *SuspendedTimesConfig) {
	if src.SearxEngineAccessDenied != 0 {
		dst.SearxEngineAccessDenied = src.SearxEngineAccessDenied
	}
	if src.SearxEngineCaptcha != 0 {
		dst.SearxEngineCaptcha = src.SearxEngineCaptcha
	}
	if src.SearxEngineTooManyRequests != 0 {
		dst.SearxEngineTooManyRequests = src.SearxEngineTooManyRequests
	}
	if src.CfSearxEngineCaptcha != 0 {
		dst.CfSearxEngineCaptcha = src.CfSearxEngineCaptcha
	}
	if src.CfSearxEngineAccessDenied != 0 {
		dst.CfSearxEngineAccessDenied = src.CfSearxEngineAccessDenied
	}
	if src.RecaptchaSearxEngineCaptcha != 0 {
		dst.RecaptchaSearxEngineCaptcha = src.RecaptchaSearxEngineCaptcha
	}
}

func overlayServer(dst *ServerConfig, src *ServerConfig) {
	if src.Port != 0 {
		dst.Port = src.Port
	}
	if src.BindAddress != "" {
		dst.BindAddress = src.BindAddress
	}
	if src.Limiter {
		dst.Limiter = true
	}
	if src.PublicInstance {
		dst.PublicInstance = true
	}
	if src.SecretKey != "" {
		dst.SecretKey = src.SecretKey
	}
	if src.BaseURL != nil {
		dst.BaseURL = src.BaseURL
	}
	if src.ImageProxy {
		dst.ImageProxy = true
	}
	if src.HTTPProtocolVersion != "" {
		dst.HTTPProtocolVersion = src.HTTPProtocolVersion
	}
	if src.Method != "" {
		dst.Method = src.Method
	}
	if len(src.DefaultHTTPHeaders) > 0 {
		if dst.DefaultHTTPHeaders == nil {
			dst.DefaultHTTPHeaders = make(map[string]string)
		}
		for k, v := range src.DefaultHTTPHeaders {
			dst.DefaultHTTPHeaders[k] = v
		}
	}
}

func overlayOutgoing(dst *OutgoingConfig, src *OutgoingConfig) {
	if src.UserAgentSuffix != "" {
		dst.UserAgentSuffix = src.UserAgentSuffix
	}
	if src.RequestTimeout != 0 {
		dst.RequestTimeout = src.RequestTimeout
	}
	if src.EnableHTTP2 {
		dst.EnableHTTP2 = true
	}
	if src.Verify != nil {
		dst.Verify = src.Verify
	}
	if src.MaxRequestTimeout != nil {
		dst.MaxRequestTimeout = src.MaxRequestTimeout
	}
	if src.PoolConnections != 0 {
		dst.PoolConnections = src.PoolConnections
	}
	if src.PoolMaxsize != 0 {
		dst.PoolMaxsize = src.PoolMaxsize
	}
	if src.KeepaliveExpiry != 0 {
		dst.KeepaliveExpiry = src.KeepaliveExpiry
	}
	if src.MaxRedirects != 0 {
		dst.MaxRedirects = src.MaxRedirects
	}
	if src.Retries != 0 {
		dst.Retries = src.Retries
	}
	if src.Proxies != nil {
		dst.Proxies = src.Proxies
	}
	if src.SourceIPs != nil {
		dst.SourceIPs = src.SourceIPs
	}
	if src.UsingTorProxy {
		dst.UsingTorProxy = true
	}
	if src.ExtraProxyTimeout != 0 {
		dst.ExtraProxyTimeout = src.ExtraProxyTimeout
	}
	if src.UserAgent != "" {
		dst.UserAgent = src.UserAgent
	}
	if src.Timeout != 0 {
		dst.RequestTimeout = float64(src.Timeout)
	}
	if src.EnableHTTP {
		dst.EnableHTTP = true
	}
	if src.Networks != nil {
		if dst.Networks == nil {
			dst.Networks = make(map[string]OutgoingNetworkOverride)
		}
		for k, v := range src.Networks {
			dst.Networks[k] = v
		}
	}
	if src.RetryOnHTTPError != nil {
		dst.RetryOnHTTPError = src.RetryOnHTTPError
	}
}

func overlayUI(dst *UIConfig, src *UIConfig) {
	if src.StaticPath != "" {
		dst.StaticPath = src.StaticPath
	}
	if src.TemplatesPath != "" {
		dst.TemplatesPath = src.TemplatesPath
	}
	if src.DefaultTheme != "" {
		dst.DefaultTheme = src.DefaultTheme
	}
	if src.DefaultLocale != "" {
		dst.DefaultLocale = src.DefaultLocale
	}
	if src.CenterAlignment {
		dst.CenterAlignment = true
	}
	if src.ResultsOnNewTab {
		dst.ResultsOnNewTab = true
	}
	if src.QueryInTitle {
		dst.QueryInTitle = true
	}
	if src.CacheURL != "" {
		dst.CacheURL = src.CacheURL
	}
	if src.SearchOnCategorySelect {
		dst.SearchOnCategorySelect = true
	}
	if src.Hotkeys != "" {
		dst.Hotkeys = src.Hotkeys
	}
	if src.URLFormatting != "" {
		dst.URLFormatting = src.URLFormatting
	}
	if src.ThemeArgs.SimpleStyle != "" {
		dst.ThemeArgs.SimpleStyle = src.ThemeArgs.SimpleStyle
	}
}

// -------- Use Default Settings --------

func applyUseDefaultSettings(cfg *Config) {
	if cfg.UseDefaultSettings == nil {
		return
	}

	// Apply engine remove/keep_only filters
	removeSet := make(map[string]bool)
	for _, name := range cfg.UseDefaultSettings.Engines.Remove {
		removeSet[strings.ToLower(name)] = true
	}
	keepSet := make(map[string]bool)
	hasKeepOnly := len(cfg.UseDefaultSettings.Engines.KeepOnly) > 0
	for _, name := range cfg.UseDefaultSettings.Engines.KeepOnly {
		keepSet[strings.ToLower(name)] = true
	}

	filtered := make([]EngineConfig, 0, len(cfg.Engines))
	for _, eng := range cfg.Engines {
		lookupName := strings.ToLower(eng.Engine)
		if lookupName == "" {
			lookupName = strings.ToLower(eng.Name)
		}

		if removeSet[lookupName] {
			continue
		}
		if hasKeepOnly && !keepSet[lookupName] {
			continue
		}
		filtered = append(filtered, eng)
	}
	cfg.Engines = filtered

	// Consume the use_default_settings block
	cfg.UseDefaultSettings = nil
}

// -------- Validate --------

func (c *Config) Validate() error {
	// Server
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port)
	}
	if c.Server.HTTPProtocolVersion != "" && !validHTTPVersions[c.Server.HTTPProtocolVersion] {
		return fmt.Errorf("server.http_protocol_version must be 1.0 or 1.1, got %q", c.Server.HTTPProtocolVersion)
	}
	if c.Server.Method != "" && !validMethods[c.Server.Method] {
		return fmt.Errorf("server.method must be GET or POST, got %q", c.Server.Method)
	}

	// Search
	if c.Search.SafeSearch < 0 || c.Search.SafeSearch > 2 {
		return fmt.Errorf("search.safe_search must be 0, 1, or 2, got %d", c.Search.SafeSearch)
	}
	if c.Search.MaxResults <= 0 {
		c.Search.MaxResults = 10
	}
	if c.Search.DefaultCategory != "" && !validCategories[c.Search.DefaultCategory] {
		return fmt.Errorf("search.default_category %q is not a valid category", c.Search.DefaultCategory)
	}

	// Engines
	engineNames := make(map[string]bool)
	engineShortcuts := make(map[string]string)
	for i, eng := range c.Engines {
		lookupName := eng.Engine
		if lookupName == "" {
			lookupName = eng.Name
		}
		if lookupName == "" {
			return fmt.Errorf("engine[%d]: name and engine are both empty", i)
		}

		key := strings.ToLower(lookupName)
		if engineNames[key] {
			return fmt.Errorf("engine[%d]: duplicate engine name %q", i, lookupName)
		}
		engineNames[key] = true

		if eng.Weight < 0 {
			return fmt.Errorf("engine[%d] (%s): weight must be >= 0, got %f", i, lookupName, eng.Weight)
		}

		// Check for empty tokens
		if len(eng.Tokens) > 0 {
			for j, tok := range eng.Tokens {
				if tok == "" {
					return fmt.Errorf("engine[%d] (%s): token[%d] is empty", i, lookupName, j)
				}
			}
		}

		for _, cat := range eng.Categories {
			if !validCategories[cat] {
				return fmt.Errorf("engine[%d] (%s): unknown category %q", i, lookupName, cat)
			}
		}

		if eng.Shortcut != "" {
			if existing, ok := engineShortcuts[eng.Shortcut]; ok {
				return fmt.Errorf("engine[%d] (%s): duplicate shortcut %q (already used by %s)", i, lookupName, eng.Shortcut, existing)
			}
			engineShortcuts[eng.Shortcut] = lookupName
		}
	}

	// CategoriesAsTabs
	for key := range c.CategoriesAsTabs {
		if !validCategories[key] {
			return fmt.Errorf("categories_as_tabs: unknown category %q", key)
		}
	}

	return nil
}

// -------- Env overrides --------

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SEARGO_DEBUG"); v != "" {
		cfg.General.Debug = v == "true" || v == "1"
	}
	if v := os.Getenv("SEARGO_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("SEARGO_SERVER_BIND_ADDRESS"); v != "" {
		cfg.Server.BindAddress = v
	}
	if v := os.Getenv("SEARGO_SERVER_SECRET_KEY"); v != "" {
		cfg.Server.SecretKey = v
	}
	if v := os.Getenv("SEARGO_SERVER_BASE_URL"); v != "" {
		cfg.Server.BaseURL = &v
	}
	if v := os.Getenv("SEARGO_VALKEY_URL"); v != "" {
		cfg.Valkey.URL = &v
	}
	// Legacy env vars
	if v := os.Getenv("SEARGO_CACHE_REDIS_ADDR"); v != "" {
		cfg.Cache.RedisAddr = v
	}
	for i := range cfg.Engines {
		envKey := fmt.Sprintf("SEARGO_ENGINE_%s_API_KEY", strings.ToUpper(cfg.Engines[i].Name))
		if v := os.Getenv(envKey); v != "" {
			cfg.Engines[i].APIKey = v
		}
	}
}
