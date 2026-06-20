package builtin

import (
	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	plugin.RegisterBuiltin("infinite_scroll", func() plugin.Plugin {
		return &infiniteScrollPlugin{}
	})
}

// infiniteScrollPlugin is a metadata-only plugin that enables infinite scroll in the UI.
// It does not perform any hooks — all hooks return true/nil.
type infiniteScrollPlugin struct{}

func (i *infiniteScrollPlugin) ID() string { return "infinite_scroll" }

func (i *infiniteScrollPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                i.ID(),
		Name:              "Infinite Scroll",
		Description:       "Enable infinite scroll for search results — automatically load the next page when scrolling to the bottom.",
		PreferenceSection: "ui",
	}
}

func (i *infiniteScrollPlugin) Init(ctx *plugin.AppContext) bool {
	return true
}

func (i *infiniteScrollPlugin) PreSearch(ctx *plugin.SearchContext) bool {
	return true
}

func (i *infiniteScrollPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool {
	return true
}

func (i *infiniteScrollPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	return nil
}
