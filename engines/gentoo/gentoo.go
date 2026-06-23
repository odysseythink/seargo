package gentoo

import (
	"context"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/engine/bases"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine.Register("gentoo", &Gentoo{})
}

// Gentoo wraps the MediaWiki base engine for wiki.gentoo.org.
type Gentoo struct {
	inner engine.Engine
}

func (g *Gentoo) Name() string { return "gentoo" }

func (g *Gentoo) Categories() []models.Category {
	return []models.Category{models.CategoryIT}
}

func (g *Gentoo) About() engine.EngineAbout {
	return engine.EngineAbout{Website: "https://wiki.gentoo.org"}
}

func (g *Gentoo) Capabilities() engine.Capabilities { return engine.Capabilities{} }

func (g *Gentoo) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }

func (g *Gentoo) Setup(cfg engine.EngineInitConfig) bool {
	baseURL := "https://wiki.gentoo.org/"
	apiPath := "api.php"
	if cfg.Extra != nil {
		if s, ok := cfg.Extra["base_url"].(string); ok && s != "" {
			baseURL = s
		}
		if s, ok := cfg.Extra["api_path"].(string); ok && s != "" {
			apiPath = s
		}
	}

	g.inner = bases.NewMediaWikiEngine("gentoo", cfg.Categories, bases.MediaWikiConfig{
		BaseURL: baseURL,
		APIPath: apiPath,
	})
	return g.inner.Setup(cfg)
}

func (g *Gentoo) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return g.inner.Search(ctx, req)
}
