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

func TestEngineTraits_Resolve_Bing(t *testing.T) {
	traits := EngineTraits{
		DataType:  "traits_v1",
		AllLocale: "clear",
		Regions: map[string]string{
			"zh-CN": "zh-cn",
			"en-US": "en-us",
			"fr-FR": "fr-fr",
		},
	}

	resolved := traits.Resolve("zh-CN")
	if resolved.Region != "zh-cn" {
		t.Errorf("zh-CN region = %q, want zh-cn", resolved.Region)
	}
	if resolved.All {
		t.Error("All should be false when region resolved")
	}

	resolved = traits.Resolve("fr-BE")
	if resolved.Region != "fr-fr" {
		t.Errorf("fr-BE region = %q, want fr-fr (territory fallback)", resolved.Region)
	}
}

func TestEngineTraits_Resolve_Wikipedia(t *testing.T) {
	traits := EngineTraits{
		Languages: map[string]string{
			"de": "de",
			"fr": "fr",
			"zh": "zh",
		},
	}

	resolved := traits.Resolve("de")
	if resolved.Language != "de" {
		t.Errorf("de language = %q, want de", resolved.Language)
	}

	// zh-CN → zh language (language fallback)
	resolved = traits.Resolve("zh-CN")
	if resolved.Language != "zh" {
		t.Errorf("zh-CN language = %q, want zh", resolved.Language)
	}
}

func TestEngineTraits_Resolve_EmptyTraits(t *testing.T) {
	traits := EngineTraits{AllLocale: "clear"}

	resolved := traits.Resolve("zh-CN")
	if !resolved.All {
		t.Error("All should be true when no languages/regions matched")
	}
}

func TestLoadGoogleTraits(t *testing.T) {
	data, err := os.ReadFile("../../data/engine_traits.json")
	require.NoError(t, err)

	var traits EngineTraitsMap
	require.NoError(t, json.Unmarshal(data, &traits))

	g, ok := traits.Lookup("google")
	require.True(t, ok, "google traits missing")
	assert.Equal(t, "traits_v1", g.DataType)
	assert.Equal(t, "ZZ", g.AllLocale)
	assert.NotEmpty(t, g.Languages["en"])
	assert.NotEmpty(t, g.Languages["zh-CN"])
	assert.NotEmpty(t, g.Regions["en-US"])
	assert.NotEmpty(t, g.Regions["zh-CN"])

	resolved := g.Resolve("en-US")
	assert.Equal(t, "lang_en", resolved.Language)
	assert.Equal(t, "US", resolved.Region)
	assert.False(t, resolved.All)

	resolvedAll := g.Resolve("all")
	assert.True(t, resolvedAll.All)
}

func TestEngineTraits_Resolve_NoMatch(t *testing.T) {
	traits := EngineTraits{
		Languages: map[string]string{"en": "en"},
	}

	resolved := traits.Resolve("xx-YY")
	if resolved.Language != "" {
		t.Errorf("unmatchable: Language = %q, want empty", resolved.Language)
	}
	if resolved.All {
		t.Error("All should be false when all_locale is empty")
	}
}
