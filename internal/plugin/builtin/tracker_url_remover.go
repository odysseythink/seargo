package builtin

import (
	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/internal/plugin/deps"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	plugin.RegisterBuiltin("tracker_url_remover", func() plugin.Plugin {
		return &trackerURLRemoverPlugin{}
	})
}

// trackerURLRemoverPlugin strips tracking parameters from result URLs.
// It runs on all results (no keywords).
type trackerURLRemoverPlugin struct{}

func (t *trackerURLRemoverPlugin) ID() string { return "tracker_url_remover" }

func (t *trackerURLRemoverPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                t.ID(),
		Name:              "Tracker URL Remover",
		Description:       "Remove tracking parameters (UTM, Facebook, Google, etc.) from result URLs.",
		PreferenceSection: "privacy",
	}
}

func (t *trackerURLRemoverPlugin) Init(ctx *plugin.AppContext) bool {
	return true
}

func (t *trackerURLRemoverPlugin) PreSearch(ctx *plugin.SearchContext) bool {
	return true
}

func (t *trackerURLRemoverPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool {
	deps.InitTrackerPatterns()

	r.FilterURLs(func(r *models.Result, field string, url string) (string, bool) {
		cleaned, _ := deps.TrackerCleanURL(url)
		return cleaned, true
	})

	return true
}

func (t *trackerURLRemoverPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	return nil
}
