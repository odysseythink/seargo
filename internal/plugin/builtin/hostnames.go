package builtin

import (
	"regexp"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
)

// hostnamesConfig is the configuration for the hostnames plugin.
type hostnamesConfig struct {
	Replace      map[*regexp.Regexp]string
	Remove       []*regexp.Regexp
	HighPriority []*regexp.Regexp
	LowPriority  []*regexp.Regexp
}

type hostnamesPlugin struct {
	config *hostnamesConfig
}

func init() {
	plugin.RegisterBuiltin("hostnames", func() plugin.Plugin { return &hostnamesPlugin{} })
}

func (p *hostnamesPlugin) ID() string { return "hostnames" }

func (p *hostnamesPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                "hostnames",
		Name:              "Hostnames",
		Description:       "Filter and rewrite search results by hostname patterns",
		PreferenceSection: "general",
	}
}

func (p *hostnamesPlugin) Init(ctx *plugin.AppContext) bool {
	cfg, ok := ctx.Config.(*hostnamesConfig)
	if !ok || cfg == nil {
		return false
	}
	p.config = cfg
	return true
}

func (p *hostnamesPlugin) PreSearch(ctx *plugin.SearchContext) bool { return true }

func (p *hostnamesPlugin) OnResult(ctx *plugin.SearchContext, result *models.Result) bool {
	if p.config == nil {
		return true
	}

	// Check Remove patterns
	for _, re := range p.config.Remove {
		if re.MatchString(result.URL) {
			return false
		}
	}

	// Check Replace patterns
	if len(p.config.Replace) > 0 {
		result.FilterURLs(func(r *models.Result, field string, url string) (string, bool) {
			for re, replacement := range p.config.Replace {
				if re.MatchString(url) {
					return re.ReplaceAllString(url, replacement), true
				}
			}
			return url, true
		})
	}

	return true
}

func (p *hostnamesPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result { return nil }
