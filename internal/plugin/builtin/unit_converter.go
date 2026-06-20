package builtin

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/internal/plugin/deps"
	"github.com/seargo/seargo/pkg/models"
)

type unitConverterPlugin struct{}

var unitConvPattern = regexp.MustCompile(`^(\d+\.?\d*)\s+(\S+)\s+to\s+(\S+)$`)

func init() {
	plugin.RegisterBuiltin("unit_converter", func() plugin.Plugin { return &unitConverterPlugin{} })
}

func (p *unitConverterPlugin) ID() string { return "unit_converter" }

func (p *unitConverterPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                "unit_converter",
		Name:              "Unit Converter",
		Description:       "Convert between different units of measurement",
		PreferenceSection: "query",
		Examples:          []string{"10 km to mi", "100 cm to in"},
	}
}

func (p *unitConverterPlugin) Init(ctx *plugin.AppContext) bool                { return true }
func (p *unitConverterPlugin) PreSearch(ctx *plugin.SearchContext) bool       { return true }
func (p *unitConverterPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool { return true }

func (p *unitConverterPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	q := strings.TrimSpace(ctx.Query)
	matches := unitConvPattern.FindStringSubmatch(q)
	if len(matches) < 4 {
		return nil
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil
	}

	fromUnit := matches[2]
	toUnit := matches[3]

	fromEntries := deps.LookupUnit(fromUnit)
	toEntries := deps.LookupUnit(toUnit)

	if len(fromEntries) == 0 || len(toEntries) == 0 {
		return nil
	}

	converted, ok := deps.Convert(value, fromEntries, toEntries)
	if !ok {
		return nil
	}

	var resultStr string
	if converted == math.Trunc(converted) && !math.IsInf(converted, 0) && !math.IsNaN(converted) {
		resultStr = fmt.Sprintf("%s = %.0f %s", matches[0], converted, toUnit)
	} else {
		resultStr = fmt.Sprintf("%s = %g %s", matches[0], converted, toUnit)
	}

	return []models.Result{{
		Kind:    "answer",
		Title:   resultStr,
		Content: fmt.Sprintf("%g", converted),
		Engine:  "unit_converter",
	}}
}
