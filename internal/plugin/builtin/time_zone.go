package builtin

import (
	"fmt"
	"strings"
	"time"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/internal/plugin/deps"
	"github.com/seargo/seargo/pkg/models"
)

type timeZonePlugin struct{}

func init() {
	plugin.RegisterBuiltin("time_zone", func() plugin.Plugin { return &timeZonePlugin{} })
}

func (p *timeZonePlugin) ID() string { return "time_zone" }

func (p *timeZonePlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                "time_zone",
		Name:              "Time Zone",
		Description:       "Show current time for a city or timezone",
		PreferenceSection: "query",
		Keywords:          []string{"time", "timezone", "now", "clock", "timezones"},
		Examples:          []string{"time Berlin", "time Tokyo", "now"},
	}
}

func (p *timeZonePlugin) Init(ctx *plugin.AppContext) bool                { return true }
func (p *timeZonePlugin) PreSearch(ctx *plugin.SearchContext) bool       { return true }
func (p *timeZonePlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool { return true }

func (p *timeZonePlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	q := strings.TrimSpace(ctx.Query)
	parts := strings.SplitN(q, " ", 2)

	// First word must be one of the keywords (keyword check happens at storage layer,
	// but we also self-check for direct invocation).
	firstWord := strings.ToLower(strings.TrimSpace(parts[0]))
	keywords := []string{"time", "timezone", "now", "clock", "timezones"}
	matched := false
	for _, kw := range keywords {
		if firstWord == kw {
			matched = true
			break
		}
	}
	if !matched {
		return nil
	}

	var loc deps.GeoLocation
	var cityName string

	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		// No city specified — show UTC
		loc = deps.GeoLocation{Name: "UTC", Timezone: "UTC"}
		cityName = "UTC"
	} else {
		city := strings.TrimSpace(parts[1])
		var found bool
		loc, found = deps.GeoLocationByQuery(city)
		if !found {
			return nil
		}
		cityName = loc.Name
	}

	now := time.Now()
	dt := deps.NewDateTime(now, loc.Timezone)

	return []models.Result{{
		Kind:    "answer",
		Title:   fmt.Sprintf("Current time in %s", cityName),
		Content: dt.Format(),
		Engine:  "time_zone",
	}}
}
