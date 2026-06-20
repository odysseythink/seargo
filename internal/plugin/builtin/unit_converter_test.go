package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/stretchr/testify/assert"
)

func TestUnitConverterPlugin_KMToMiles(t *testing.T) {
	p := &unitConverterPlugin{}
	ctx := &plugin.SearchContext{Query: "10 km to mi"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, "10 km to mi")
	assert.Contains(t, results[0].Title, "mi")
	assert.Equal(t, "unit_converter", results[0].Engine)
}

func TestUnitConverterPlugin_InToCM(t *testing.T) {
	p := &unitConverterPlugin{}
	ctx := &plugin.SearchContext{Query: "12 in to cm"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, "12 in to cm")
	assert.Contains(t, results[0].Title, "cm")
}

func TestUnitConverterPlugin_NoMatch(t *testing.T) {
	p := &unitConverterPlugin{}
	ctx := &plugin.SearchContext{Query: "regular search query"}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestUnitConverterPlugin_KGToLB(t *testing.T) {
	p := &unitConverterPlugin{}
	ctx := &plugin.SearchContext{Query: "5 kg to lb"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, "5 kg to lb")
}

func TestUnitConverterPlugin_InvalidUnits(t *testing.T) {
	p := &unitConverterPlugin{}
	ctx := &plugin.SearchContext{Query: "10 xyz to abc"}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestUnitConverterPlugin_Decimal(t *testing.T) {
	p := &unitConverterPlugin{}
	ctx := &plugin.SearchContext{Query: "3.5 km to mi"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, "3.5 km to mi")
}
