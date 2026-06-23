package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/engine"
)

func TestParseGoogleLanguages(t *testing.T) {
	html := `<html><body>
<select name="hl"><option value="en">English</option><option value="zh-CN">中文</option></select>
</body></html>`
	got, err := parseGoogleLanguages(strings.NewReader(html))
	require.NoError(t, err)
	assert.Equal(t, "lang_en", got["en"])
	assert.Equal(t, "lang_zh-CN", got["zh-CN"])
}

func TestParseGoogleRegions_SkipsKnownCodes(t *testing.T) {
	html := `<html><body>
<select name="gl"><option value="US">United States</option><option value="UN">Unspecified</option><option value="DE">Germany</option></select>
</body></html>`
	got, err := parseGoogleRegions(strings.NewReader(html))
	require.NoError(t, err)
	assert.Equal(t, "US", got["US"])
	assert.Equal(t, "DE", got["DE"])
	assert.NotContains(t, got, "UN")
}

func TestParseSupportedDomains(t *testing.T) {
	input := `.google.com
.google.de
.google.com.hk
`
	got, err := parseSupportedDomains(strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, "www.google.com", got["COM"])
	assert.Equal(t, "www.google.de", got["DE"])
	assert.Equal(t, "www.google.com.hk", got["HK"])
}

func TestMergeGoogleTraits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "engine_traits.json")

	// Pre-existing unrelated traits must be preserved.
	require.NoError(t, os.WriteFile(path, []byte(`{"duckduckgo":{"data_type":"traits_v1"}}`), 0644))

	traits := engine.EngineTraits{
		DataType:  "traits_v1",
		Languages: map[string]string{"en": "lang_en"},
		Regions:   map[string]string{"en-US": "US"},
		AllLocale: "ZZ",
	}
	require.NoError(t, mergeGoogleTraits(path, traits))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"google"`)
	assert.Contains(t, string(data), `"duckduckgo"`)
}
