package bases

import (
	"context"
	"fmt"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/pkg/models"
)

// CommandConfig defines the configuration for a command-line based engine.
type CommandConfig struct {
	// TODO: Offline — not yet implemented (deferred per scope)
}

type commandEngine struct {
	name       string
	categories []models.Category
	cfg        CommandConfig
}

// NewCommandEngine creates a command-line based engine.
// Returns an engine that reports Setup failure (not yet implemented).
func NewCommandEngine(name string, categories []models.Category, cfg CommandConfig) engine.Engine {
	return &commandEngine{
		name:       name,
		categories: categories,
		cfg:        cfg,
	}
}

func (e *commandEngine) Name() string                     { return e.name }
func (e *commandEngine) Categories() []models.Category     { return e.categories }
func (e *commandEngine) About() engine.EngineAbout         { return engine.EngineAbout{} }
func (e *commandEngine) Capabilities() engine.Capabilities { return engine.Capabilities{} }

func (e *commandEngine) Setup(cfg engine.EngineInitConfig) bool {
	return false // Not yet implemented
}

func (e *commandEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return false
}

func (e *commandEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return nil, fmt.Errorf("command engine %s: not yet implemented", e.name)
}
