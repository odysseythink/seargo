package plugin

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGobRoundTrip_SearchContextAndResult(t *testing.T) {
	now := time.Now()
	original := &SearchContext{
		Query:       "hello world",
		RawQuery:    "!general hello world",
		Lang:        "en",
		Locale:      "en-US",
		SafeSearch:  1,
		PageNo:      2,
		TimeRange:   "year",
		RemoteAddr:  "127.0.0.1",
		UserPlugins: []string{"echo", "filter"},
		Preferences: map[string]any{
			"theme":   "dark",
			"timeout": 30,
		},
	}
	result := &models.Result{
		Kind:         "answer",
		Title:        "Example",
		URL:          "https://example.com",
		Content:      "content",
		Engine:       "example",
		Engines:      []string{"example"},
		Category:     models.CategoryGeneral,
		Score:        1.5,
		Positions:    []int{1, 2},
		ThumbnailURL: "https://example.com/t.jpg",
		PublishedAt:  &now,
		Domain:       "example.com",
		Favicon:      "https://example.com/f.ico",
		IsOnion:      true,
		Extra: map[string]any{
			"key": "value",
			"num": 42,
		},
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	require.NoError(t, enc.Encode(original))
	require.NoError(t, enc.Encode(result))
	require.NoError(t, enc.Encode([]models.Result{*result}))

	dec := gob.NewDecoder(&buf)
	var decodedCtx SearchContext
	var decodedResult models.Result
	var decodedResults []models.Result
	require.NoError(t, dec.Decode(&decodedCtx))
	require.NoError(t, dec.Decode(&decodedResult))
	require.NoError(t, dec.Decode(&decodedResults))

	assert.Equal(t, original.Query, decodedCtx.Query)
	assert.Equal(t, original.Preferences, decodedCtx.Preferences)
	assert.Equal(t, result.Title, decodedResult.Title)
	assert.Equal(t, result.Positions, decodedResult.Positions)
	assert.Equal(t, result.Extra, decodedResult.Extra)
	assert.Len(t, decodedResults, 1)
	assert.Equal(t, result.Title, decodedResults[0].Title)
}

func TestHandshakeConfig(t *testing.T) {
	assert.Equal(t, uint(1), HandshakeConfig.ProtocolVersion)
	assert.Equal(t, "SEARGO_PLUGIN", HandshakeConfig.MagicCookieKey)
	assert.Equal(t, "seargo-external-plugin", HandshakeConfig.MagicCookieValue)
}
