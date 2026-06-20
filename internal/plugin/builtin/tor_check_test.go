package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/stretchr/testify/assert"
)

func TestTorCheck_KeywordTriggersPostSearch(t *testing.T) {
	p := &torCheckPlugin{}

	ctx := &plugin.SearchContext{
		Query:       "tor-check",
		Preferences: map[string]any{"remote_addr": "1.2.3.4"},
		PageNo:      1,
	}

	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "answer", results[0].Kind)
	// HTTP fetch fails in test environment
	assert.Contains(t, results[0].Title, "Tor check unavailable")
}

func TestTorCheck_NoMatchReturnsNil(t *testing.T) {
	p := &torCheckPlugin{}

	ctx := &plugin.SearchContext{
		Query:       "something else",
		Preferences: map[string]any{"remote_addr": "1.2.3.4"},
		PageNo:      1,
	}

	results := p.PostSearch(ctx)
	assert.Nil(t, results)
}

func TestTorCheck_PageGreaterThanOneReturnsNil(t *testing.T) {
	p := &torCheckPlugin{}

	ctx := &plugin.SearchContext{
		Query:       "tor-check",
		Preferences: map[string]any{"remote_addr": "1.2.3.4"},
		PageNo:      2,
	}

	results := p.PostSearch(ctx)
	assert.Nil(t, results)
}

func TestTorCheck_KeywordsList(t *testing.T) {
	p := &torCheckPlugin{}
	keywords := p.Info().Keywords

	expected := []string{"tor-check", "tor_check", "torcheck", "tor", "tor check"}
	assert.ElementsMatch(t, expected, keywords)
}

func TestTorCheck_UnknownIP(t *testing.T) {
	p := &torCheckPlugin{}

	ctx := &plugin.SearchContext{
		Query:       "tor-check",
		Preferences: map[string]any{},
		PageNo:      1,
	}

	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, "Tor check unavailable")
}

func TestTorCheck_AllKeywordsTrigger(t *testing.T) {
	for _, kw := range []string{"tor-check", "tor_check", "torcheck", "tor", "tor check"} {
		t.Run(kw, func(t *testing.T) {
			p := &torCheckPlugin{}
			ctx := &plugin.SearchContext{
				Query:       kw,
				Preferences: map[string]any{"remote_addr": "10.0.0.1"},
				PageNo:      1,
			}
			results := p.PostSearch(ctx)
			assert.Len(t, results, 1, "keyword %q should trigger", kw)
		})
	}
}
