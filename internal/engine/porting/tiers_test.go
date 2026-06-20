package porting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTiers_NotEmpty(t *testing.T) {
	tiers := AllTiers()
	assert.NotEmpty(t, tiers)
}

func TestTiers_SortedByPriority(t *testing.T) {
	tiers := AllTiers()
	for i := 1; i < len(tiers); i++ {
		assert.LessOrEqual(t, tiers[i-1].Priority, tiers[i].Priority,
			"tiers must be sorted by ascending priority")
	}
}

func TestTiers_TotalEngines(t *testing.T) {
	total := TotalEngines()
	assert.Greater(t, total, 200, "should track at least 200 engines")
}

func TestTier1_ContainsMajorEngines(t *testing.T) {
	t1 := Tier1()
	names := make(map[string]bool)
	for _, e := range t1 {
		names[e.Name] = true
	}
	assert.True(t, names["google"], "Tier 1 must include google")
	assert.True(t, names["bing"], "Tier 1 must include bing")
	assert.True(t, names["wikipedia"], "Tier 1 must include wikipedia")
}
