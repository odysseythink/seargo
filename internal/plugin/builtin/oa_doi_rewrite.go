package builtin

import (
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/internal/plugin/deps"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	plugin.RegisterBuiltin("oa_doi_rewrite", func() plugin.Plugin {
		return &oaDOIRewritePlugin{}
	})
}

// oaDOIRewritePlugin rewrites result URLs to use a configured open-access DOI resolver.
// It runs on all results (no keywords).
type oaDOIRewritePlugin struct {
	preferredResolver string
	resolvers         map[string]string
	defaultResolver   string
}

func (o *oaDOIRewritePlugin) ID() string { return "oa_doi_rewrite" }

func (o *oaDOIRewritePlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                o.ID(),
		Name:              "OA DOI Rewrite",
		Description:       "Rewrite DOI links to use your preferred open-access resolver (e.g., Sci-Hub, Unpaywall).",
		PreferenceSection: "general",
	}
}

var defaultDOIResolvers = map[string]string{
	"oadoi.org": "https://oadoi.org/",
	"doi.org":   "https://doi.org/",
	"scihub":    "https://sci-hub.se/",
}

func (o *oaDOIRewritePlugin) Init(ctx *plugin.AppContext) bool {
	o.defaultResolver = "oadoi.org"

	cfg, ok := ctx.Config.(*config.Config)
	if !ok || cfg == nil {
		o.preferredResolver = o.defaultResolver
		o.resolvers = defaultDOIResolvers
		return true
	}

	preferred := cfg.DefaultDOIResolver
	if preferred == "" {
		preferred = o.defaultResolver
	}

	resolvers := cfg.DOIRsolvers
	if len(resolvers) == 0 {
		resolvers = defaultDOIResolvers
	}

	o.preferredResolver = preferred
	o.resolvers = resolvers
	return true
}

func (o *oaDOIRewritePlugin) PreSearch(ctx *plugin.SearchContext) bool {
	return true
}

func (o *oaDOIRewritePlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool {
	doi, ok := deps.ExtractDOI(r.URL)
	if !ok {
		return true
	}

	resolver := deps.GetDOIResolverURL(o.preferredResolver, o.resolvers, o.defaultResolver)
	r.URL = resolver + doi

	return true
}

func (o *oaDOIRewritePlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	return nil
}
