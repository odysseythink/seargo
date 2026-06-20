package engine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineTraits_LoadFromJSON(t *testing.T) {
	data, err := os.ReadFile("../../data/engine_traits.json")
	require.NoError(t, err)

	var tm EngineTraitsMap
	err = json.Unmarshal(data, &tm)
	require.NoError(t, err)

	traits, ok := tm["duckduckgo"]
	assert.True(t, ok, "duckduckgo not found in traits")
	assert.Equal(t, "traits_v1", traits.DataType)
	assert.NotEmpty(t, traits.Languages)
}

func TestEngineTraits_Resolve(t *testing.T) {
	traits := EngineTraits{
		Languages: map[string]string{"en": "en-US", "zh": "zh-CN", "fr": "fr"},
		Regions:   map[string]string{"us": "en-US", "cn": "zh-CN"},
		DataType:  "traits_v1",
	}

	// No config filter — all languages pass through
	resolved := resolveTraits(traits, "", "")
	assert.Len(t, resolved.Languages, 3)
	assert.Len(t, resolved.Regions, 2)

	// With config language filter
	resolved = resolveTraits(traits, "zh", "")
	assert.Len(t, resolved.Languages, 1)
	_, ok := resolved.Languages["zh"]
	assert.True(t, ok)

	// With config region filter
	resolved = resolveTraits(traits, "", "cn")
	assert.Len(t, resolved.Regions, 1)
	_, ok = resolved.Regions["cn"]
	assert.True(t, ok)
}

func TestEngineTraits_EmptyMaps(t *testing.T) {
	traits := resolveTraits(EngineTraits{}, "", "")
	assert.NotNil(t, traits.Languages)
	assert.NotNil(t, traits.Regions)
	assert.Empty(t, traits.Languages)
}

func TestEngineTraits_LookupByName(t *testing.T) {
	tm := EngineTraitsMap{
		"google": {Languages: map[string]string{"en": "en"}},
	}

	traits, ok := tm.Lookup("google")
	assert.True(t, ok)
	assert.Len(t, traits.Languages, 1)

	// Unknown engine returns empty
	traits, ok = tm.Lookup("nonexistent")
	assert.False(t, ok)
	assert.Empty(t, traits.Languages)
}
