package builtin

import (
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

func (o *oaDOIRewritePlugin) Init(ctx *plugin.AppContext) bool {
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
