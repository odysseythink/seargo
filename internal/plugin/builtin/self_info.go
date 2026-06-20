package builtin

import (
	"fmt"
	"strings"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
)

type selfInfoPlugin struct{}

func init() {
	plugin.RegisterBuiltin("self_info", func() plugin.Plugin { return &selfInfoPlugin{} })
}

func (p *selfInfoPlugin) ID() string { return "self_info" }

func (p *selfInfoPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                "self_info",
		Name:              "Self Info",
		Description:       "Show your own IP address or user agent",
		PreferenceSection: "general",
		Keywords:          []string{"ip", "user-agent"},
		Examples:          []string{"ip", "user-agent"},
	}
}

func (p *selfInfoPlugin) Init(ctx *plugin.AppContext) bool                { return true }
func (p *selfInfoPlugin) PreSearch(ctx *plugin.SearchContext) bool       { return true }
func (p *selfInfoPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool { return true }

func (p *selfInfoPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	q := strings.TrimSpace(strings.ToLower(ctx.Query))

	switch q {
	case "ip":
		remoteAddr, _ := ctx.Preferences["remote_addr"].(string)
		if remoteAddr == "" {
			return nil
		}
		return []models.Result{{
			Kind:    "answer",
			Title:   fmt.Sprintf("Your IP address is %s", remoteAddr),
			Content: remoteAddr,
			Engine:  "self_info",
		}}
	case "user-agent":
		userAgent, _ := ctx.Preferences["user_agent"].(string)
		if userAgent == "" {
			return nil
		}
		return []models.Result{{
			Kind:    "answer",
			Title:   fmt.Sprintf("Your user agent is %s", userAgent),
			Content: userAgent,
			Engine:  "self_info",
		}}
	}

	return nil
}
