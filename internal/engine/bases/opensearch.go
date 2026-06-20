package bases

import (
	"context"
	"fmt"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/pkg/models"
)

// OpenSearchConfig defines the configuration for an OpenSearch descriptor engine.
type OpenSearchConfig struct {
	// TODO: Offline — not yet implemented (deferred per scope)
}

type openSearchEngine struct {
	name       string
	categories []models.Category
	cfg        OpenSearchConfig
}

// NewOpenSearchEngine creates an OpenSearch descriptor based engine.
// Returns an engine that reports Setup failure (not yet implemented).
func NewOpenSearchEngine(name string, categories []models.Category, cfg OpenSearchConfig) engine.Engine {
	return &openSearchEngine{
		name:       name,
		categories: categories,
		cfg:        cfg,
	}
}

func (e *openSearchEngine) Name() string                     { return e.name }
func (e *openSearchEngine) Categories() []models.Category     { return e.categories }
func (e *openSearchEngine) About() engine.EngineAbout         { return engine.EngineAbout{} }
func (e *openSearchEngine) Capabilities() engine.Capabilities { return engine.Capabilities{} }

func (e *openSearchEngine) Setup(cfg engine.EngineInitConfig) bool {
	return false // Not yet implemented
}

func (e *openSearchEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return false
}

func (e *openSearchEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return nil, fmt.Errorf("opensearch engine %s: not yet implemented", e.name)
}
