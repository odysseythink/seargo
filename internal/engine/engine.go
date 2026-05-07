package engine

import (
	"context"

	"github.com/seargo/seargo/pkg/models"
)

// Engine is the interface that all search engines must implement.
type Engine interface {
	Name() string
	Categories() []models.Category
	Capabilities() Capabilities
	Init(cfg map[string]any) error
	Search(ctx context.Context, req *models.Request) (*models.Response, error)
}

// Capabilities describes what features an engine supports.
type Capabilities struct {
	SupportsSafeSearch bool `json:"supports_safe_search"`
	SupportsLanguage   bool `json:"supports_language"`
	SupportsTimeRange  bool `json:"supports_time_range"`
	SupportsPagination bool `json:"supports_pagination"`
	RequiresAPIKey     bool `json:"requires_api_key"`
}

// Info describes an engine for API responses.
type Info struct {
	Name         string       `json:"name"`
	Categories   []string     `json:"categories"`
	Capabilities Capabilities `json:"capabilities"`
	Enabled      bool         `json:"enabled"`
}
