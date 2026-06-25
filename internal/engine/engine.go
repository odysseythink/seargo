package engine

import (
	"context"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

// EngineInitConfig holds per-engine runtime configuration from the config file.
type EngineInitConfig struct {
	Name       string            // display name from config
	EngineType string            // engine type key (e.g., "stackexchange" for "stackoverflow")
	Shortcut   string            // shortcut from config
	Categories []models.Category // per-engine categories (overrides defaults)
	Timeout    float64           // per-engine timeout in seconds
	Extra      map[string]any    // arbitrary extra config

	// SearXNG-compatible fields
	Paging            bool         // whether engine supports pagination
	TimeRangeSupport  bool         // whether engine supports time_range
	LanguageSupport   bool         // whether engine supports language parameter
	SafeSearch        bool         // whether engine supports safesearch
	Weight            float64      // engine weight for scoring
	DisplayErrorMsgs  bool         // show error messages to user
	EnableHTTP        bool         // allow HTTP (not just HTTPS)
	Inactive          bool         // engine inactive (skip entirely)
	Disabled          bool         // engine disabled by config
	Tokens            []string     // per-engine API tokens
	Network           string       // named network for outbound requests
	SoftMaxRedirects  int          // max redirects before soft error
	NoResultForHTTPStatus []int    // HTTP statuses treated as "no result"
	RaiseForHTTPError interface{}  // nil|bool|int|[]int for retry-on-http-error
	EngineTraits      EngineTraits // resolved language/region traits
	GoogleParams      config.GoogleEngineParams // Google-specific engine params

	// Client is the shared HTTP client for the engine to use.
	Client *httpx.Client
}

// Engine is the interface that all search engines must implement.
type Engine interface {
	Name() string
	Categories() []models.Category
	Capabilities() Capabilities
	About() EngineAbout
	Init(ctx context.Context, cfg EngineInitConfig) bool
	Setup(cfg EngineInitConfig) bool
	Search(ctx context.Context, req *models.Request) (*models.Response, error)
}

// Capabilities describes what features an engine supports.
type Capabilities struct {
	SupportsSafeSearch bool   `json:"supports_safe_search"`
	SupportsLanguage   bool   `json:"supports_language"`
	SupportsTimeRange  bool   `json:"supports_time_range"`
	SupportsPagination bool   `json:"supports_pagination"`
	RequiresAPIKey     bool   `json:"requires_api_key"`
	Shortcut           string `json:"shortcut"`
}

// EngineAbout holds descriptive metadata for an engine.
type EngineAbout struct {
	Website     string `json:"website,omitempty"`
	WikidataID  string `json:"wikidata_id,omitempty"`
	UseAPK      string `json:"use_api_key,omitempty"`
	ResultsHTML string `json:"results_html,omitempty"`
}

// Info describes an engine for API responses.
type Info struct {
	Name         string       `json:"name"`
	Categories   []string     `json:"categories"`
	Shortcut     string       `json:"shortcut"`
	Capabilities Capabilities `json:"capabilities"`
	Enabled      bool         `json:"enabled"`
}


