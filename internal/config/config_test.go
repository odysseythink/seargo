package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLoad(t *testing.T) {
	cfg, err := Load("../../configs/settings.yml")
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "general", cfg.Search.DefaultCategory)
}

func TestValidate(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080},
		Search: SearchConfig{MaxResults: 0},
	}
	err := cfg.Validate()
	require.NoError(t, err)
	assert.Equal(t, 10, cfg.Search.MaxResults)
}

func TestValidateBadPort(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Server.Port = 0
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port")

	cfg.Server.Port = 70000
	err = cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestValidateSafeSearch(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Search.SafeSearch = 3
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "safe_search")
}

func TestValidateDuplicateEngineNames(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Engines = []EngineConfig{
		{Name: "google", Engine: "google"},
		{Name: "google", Engine: "google-alt"},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestValidateDuplicateShortcuts(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Engines = []EngineConfig{
		{Name: "google", Engine: "google", Shortcut: "g"},
		{Name: "github", Engine: "github", Shortcut: "g"},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shortcut")
}

func TestValidateNegativeWeight(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Engines = []EngineConfig{
		{Name: "google", Engine: "google", Weight: -1.0},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "weight")
}

func TestValidateUnknownCategory(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Engines = []EngineConfig{
		{Name: "google", Engine: "google", Categories: []string{"general", "nonexistent"}},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "category")
}

func TestValidateHTTPProtocolVersion(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Server.HTTPProtocolVersion = "2.0"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http_protocol_version")
}

func TestValidateMethod(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Server.Method = "PUT"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "method")
}

func TestEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yml")
	content := `
server:
  port: 8080
search:
  max_results: 10
engines:
  - name: google
    engine: google
    enabled: true
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	os.Setenv("SEARGO_SERVER_SECRET_KEY", "my-secret")
	defer os.Unsetenv("SEARGO_SERVER_SECRET_KEY")

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "my-secret", cfg.Server.SecretKey)
}

func TestLayeredLoading(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "user.yml")
	content := `
general:
  instance_name: "MyInstance"
server:
  port: 9090
search:
  safe_search: 2
  default_lang: "en"
engines:
  - name: google
    engine: google
    categories: [general]
    weight: 1.0
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// User value wins
	assert.Equal(t, "MyInstance", cfg.General.InstanceName)
	// Default preserved (user did not set debug)
	assert.Equal(t, false, cfg.General.Debug)
	// User value wins
	assert.Equal(t, 9090, cfg.Server.Port)
	// Default preserved
	assert.Equal(t, "127.0.0.1", cfg.Server.BindAddress)
	// User value wins
	assert.Equal(t, 2, cfg.Search.SafeSearch)
	// Default preserved
	assert.Equal(t, 10, cfg.Search.MaxResults)
	// User engines replace defaults
	assert.Len(t, cfg.Engines, 1)
	assert.Equal(t, "google", cfg.Engines[0].Engine)
}

func TestUseDefaultSettingsRemove(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "user.yml")
	content := `
use_default_settings:
  engines:
    remove:
      - bing
      - yahoo
engines:
  - name: google
    engine: google
    categories: [general]
    weight: 1.0
  - name: bing
    engine: bing
    categories: [general]
    weight: 0.8
  - name: yahoo
    engine: yahoo
    categories: [general]
    weight: 0.7
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// After remove, only google should remain
	require.Equal(t, 1, len(cfg.Engines))
	assert.Equal(t, "google", cfg.Engines[0].Engine)
	// UseDefaultSettings should be consumed (nil)
	assert.Nil(t, cfg.UseDefaultSettings)
}

func TestUseDefaultSettingsKeepOnly(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "user.yml")
	content := `
use_default_settings:
  engines:
    keep_only:
      - google
      - wikipedia
engines:
  - name: google
    engine: google
    categories: [general]
  - name: bing
    engine: bing
    categories: [general]
  - name: wikipedia
    engine: wikipedia
    categories: [general]
  - name: yahoo
    engine: yahoo
    categories: [general]
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// After keep_only, only google and wikipedia should remain
	require.Equal(t, 2, len(cfg.Engines))
	names := []string{cfg.Engines[0].Engine, cfg.Engines[1].Engine}
	assert.Contains(t, names, "google")
	assert.Contains(t, names, "wikipedia")
}

func TestLoadTableDriven(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		check   func(t *testing.T, cfg *Config)
	}{
		{
			name: "minimal valid config",
			yaml: `
server:
  port: 8080
search:
  max_results: 10
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 8080, cfg.Server.Port)
				assert.Equal(t, 10, cfg.Search.MaxResults)
				assert.Equal(t, "127.0.0.1", cfg.Server.BindAddress)
				assert.Equal(t, "simple", cfg.UI.DefaultTheme)
			},
		},
		{
			name: "full config with all blocks",
			yaml: `
general:
  instance_name: "MySearGo"
  debug: true
search:
  safe_search: 2
  default_lang: "fr"
  languages: ["fr", "en"]
  max_results: 20
server:
  port: 9090
  bind_address: "127.0.0.1"
  http_protocol_version: "1.1"
ui:
  default_theme: "dark"
  hotkeys: "vim"
preferences:
  lock: ["language", "theme"]
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "MySearGo", cfg.General.InstanceName)
				assert.True(t, cfg.General.Debug)
				assert.Equal(t, 2, cfg.Search.SafeSearch)
				assert.Equal(t, "fr", cfg.Search.DefaultLang)
				assert.Equal(t, []string{"fr", "en"}, cfg.Search.Languages)
				assert.Equal(t, "1.1", cfg.Server.HTTPProtocolVersion)
				assert.Equal(t, "dark", cfg.UI.DefaultTheme)
				assert.Equal(t, "vim", cfg.UI.Hotkeys)
				assert.Equal(t, []string{"language", "theme"}, cfg.Preferences.Lock)
			},
		},
		{
			name: "invalid port",
			yaml: `
server:
  port: 70000
`,
			wantErr: true,
		},
		{
			name: "invalid safesearch",
			yaml: `
server:
  port: 8080
search:
  safe_search: 5
`,
			wantErr: true,
		},
		{
			name: "duplicate engine name",
			yaml: `
server:
  port: 8080
engines:
  - name: google
    engine: google
  - name: google
    engine: google
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yml")
			require.NoError(t, os.WriteFile(configPath, []byte(tt.yaml), 0644))

			cfg, err := Load(configPath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestEngineConfigNewFields(t *testing.T) {
	yamlData := `
engines:
  - name: google
    engine: google
    paging: true
    time_range_support: true
    language_support: true
    safesearch: true
    weight: 1.5
    display_error_messages: true
    enable_http: false
    inactive: false
    disabled: false
    tokens: ["token1", "token2"]
    network: "google_net"
    short_cut: g
    categories: [general, images]
    soft_max_redirects: 5
    no_result_for_http_status: [403, 404]
    raise_for_http_error: [429, 503]
`
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yamlData), cfg)
	require.NoError(t, err)
	require.Len(t, cfg.Engines, 1)

	e := cfg.Engines[0]
	assert.Equal(t, "google", e.Name)
	assert.True(t, e.Paging)
	assert.True(t, e.TimeRangeSupport)
	assert.True(t, e.LanguageSupport)
	assert.True(t, e.SafeSearch)
	assert.Equal(t, 1.5, e.Weight)
	assert.True(t, e.DisplayErrorMessages)
	assert.False(t, e.EnableHTTP)
	assert.False(t, e.Inactive)
	assert.False(t, e.Disabled)
	assert.Equal(t, []string{"token1", "token2"}, e.Tokens)
	assert.Equal(t, "google_net", e.Network)
	assert.Equal(t, 5, e.SoftMaxRedirects)
	assert.Equal(t, []int{403, 404}, e.NoResultForHTTPStatus)
}

func TestEngineConfigValidation_WeightNegative(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Port: 8080}, Search: SearchConfig{SafeSearch: 1}}
	cfg.Engines = []EngineConfig{
		{Name: "test", Engine: "test", Weight: -1},
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "weight")
}

func TestGoogleParams_Load(t *testing.T) {
	yamlData := `
engines:
  - name: google
    engine: google
    google_params:
      use_mobile_ui: true
      extra_params:
        - "hl=en"
        - "safe=off"
      consent_cookie: "YES+1"
`
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yamlData), cfg)
	require.NoError(t, err)
	require.Len(t, cfg.Engines, 1)

	g := cfg.Engines[0].GoogleParams
	assert.True(t, g.UseMobileUI)
	assert.Equal(t, []string{"hl=en", "safe=off"}, g.ExtraParams)
	assert.Equal(t, "YES+1", g.ConsentCookie)
}

func TestEngineConfigValidation_TokenEmpty(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Port: 8080}, Search: SearchConfig{SafeSearch: 1}}
	cfg.Engines = []EngineConfig{
		{Name: "test", Engine: "test", Tokens: []string{""}},
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}

func TestTrustedProxiesDefault(t *testing.T) {
	cfg := builtInDefaults()
	if len(cfg.Server.TrustedProxies) != 2 {
		t.Fatalf("expected 2 default trusted proxies, got %d", len(cfg.Server.TrustedProxies))
	}
	if cfg.Server.TrustedProxies[0] != "127.0.0.0/8" {
		t.Fatalf("expected 127.0.0.0/8 as first trusted proxy, got %q", cfg.Server.TrustedProxies[0])
	}
	if cfg.Server.TrustedProxies[1] != "::1/128" {
		t.Fatalf("expected ::1/128 as second trusted proxy, got %q", cfg.Server.TrustedProxies[1])
	}
}

func TestLoadLimiterConfig(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/limiter.toml"
	content := `[ip_lists]
block_ip = ["93.184.216.34"]
pass_ip = ["8.8.8.8"]
pass_searxng_org = true

[ip_limit]
filter_link_local = true
link_token = false

[windows]
burst_duration = "20s"
burst_max = 15
long_duration = "10m"
long_max = 150
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadLimiterConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.IPLists.BlockIP) != 1 || cfg.IPLists.BlockIP[0] != "93.184.216.34" {
		t.Fatalf("block_ip: %v", cfg.IPLists.BlockIP)
	}
	if !cfg.IPLists.PassSearxngOrg {
		t.Fatal("pass_searxng_org should be true")
	}
	if !cfg.IPLimit.FilterLinkLocal {
		t.Fatal("filter_link_local should be true")
	}
	if cfg.Windows.BurstMax != 15 {
		t.Fatalf("burst_max: got %d, want 15", cfg.Windows.BurstMax)
	}
}

func TestLoadFaviconConfig(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/favicons.toml"
	content := `
[proxy]
max_age = "7d"
resolver_timeout = "5s"
favicon_path = "data/favicon.svg"
favicon_mime_type = "image/svg+xml"

[proxy.resolver_map]
allesedv = "allesedv"
duckduckgo = "duckduckgo"
google = "google"
yandex = "yandex"

[cache]
hold_time = "30d"
blob_max_bytes = 20480
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFaviconConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Proxy.MaxAge != "7d" {
		t.Fatalf("max_age: got %v, want 7d", cfg.Proxy.MaxAge)
	}
	if len(cfg.Proxy.ResolverMap) != 4 {
		t.Fatalf("resolver_map: got %d entries, want 4", len(cfg.Proxy.ResolverMap))
	}
	if cfg.Cache.BlobMaxBytes != 20480 {
		t.Fatalf("blob_max_bytes: got %d, want 20480", cfg.Cache.BlobMaxBytes)
	}
}

func TestValidateCategoryCode(t *testing.T) {
	cfg := builtInDefaults()
	cfg.Search.DefaultCategory = "code"
	cfg.Engines = []EngineConfig{
		{Name: "github_code", Engine: "github_code", Categories: []string{"code"}},
	}
	err := cfg.Validate()
	require.NoError(t, err)
}

func TestLoad_P2Engines(t *testing.T) {
	cfg, err := Load("../../configs/settings.yml")
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range cfg.Engines {
		names[e.Name] = true
	}

	p2 := []string{
		"docker_hub", "hoogle", "mdn", "mankier",
		"openairedatasets", "openairepublications",
		"stackoverflow", "askubuntu", "superuser",
		"github_code", "gentoo", "wikicommons_files",
	}
	for _, name := range p2 {
		assert.True(t, names[name], "P2 engine %q should be in default config", name)
	}

	// github_code 使用新增的 code 分类
	for _, e := range cfg.Engines {
		if e.Name == "github_code" {
			assert.Contains(t, e.Categories, "code")
		}
	}
}
