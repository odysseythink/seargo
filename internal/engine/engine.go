package engine

import (
	"context"

	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

// EngineInitConfig holds per-engine configuration from the config file.
type EngineInitConfig struct {
	Name       string            // display name from config
	Shortcut   string            // shortcut from config
	Categories []models.Category // per-engine categories (overrides defaults)
	Timeout    float64           // per-engine timeout in seconds
	Extra      map[string]any    // arbitrary extra config
}

// Engine is the interface that all search engines must implement.
type Engine interface {
	Name() string
	Categories() []models.Category
	Capabilities() Capabilities
	Init(client *httpx.Client, cfg EngineInitConfig) error
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

// Info describes an engine for API responses.
type Info struct {
	Name         string       `json:"name"`
	Categories   []string     `json:"categories"`
	Shortcut     string       `json:"shortcut"`
	Capabilities Capabilities `json:"capabilities"`
	Enabled      bool         `json:"enabled"`
}
