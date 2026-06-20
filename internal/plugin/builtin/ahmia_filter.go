package builtin

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"strings"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/internal/plugin/deps"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	plugin.RegisterBuiltin("ahmia_filter", func() plugin.Plugin {
		return &ahmiaFilterPlugin{
			blacklist: deps.NewAhmiaBlacklist(),
		}
	})
}

// ahmiaFilterPlugin removes results from blacklisted .onion services.
// It runs on all results (no keywords) but only acts on onion results.
type ahmiaFilterPlugin struct {
	blacklist  *deps.AhmiaBlacklist
	torEnabled bool
}

func (a *ahmiaFilterPlugin) ID() string { return "ahmia_filter" }

func (a *ahmiaFilterPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                a.ID(),
		Name:              "Ahmia Filter",
		Description:       "Filter results from blacklisted .onion services. Requires Tor to be enabled.",
		PreferenceSection: "privacy",
	}
}

func (a *ahmiaFilterPlugin) Init(ctx *plugin.AppContext) bool {
	return a.torEnabled
}

func (a *ahmiaFilterPlugin) PreSearch(ctx *plugin.SearchContext) bool {
	return true
}

func (a *ahmiaFilterPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool {
	if !r.IsOnion {
		return true
	}

	host := extractHost(r.URL)
	if host == "" {
		return true
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(host)))
	if a.blacklist.Contains(hash) {
		return false
	}

	return true
}

func (a *ahmiaFilterPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	return nil
}

// extractHost parses the hostname from a URL string.
func extractHost(rawURL string) string {
	// Prepend scheme if missing for proper parsing
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
