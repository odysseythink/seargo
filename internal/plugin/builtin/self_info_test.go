package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/stretchr/testify/assert"
)

func TestSelfInfoPlugin_IP(t *testing.T) {
	p := &selfInfoPlugin{}
	ctx := &plugin.SearchContext{
		Query:       "ip",
		Preferences: map[string]any{"remote_addr": "192.168.1.1"},
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "Your IP address is 192.168.1.1", results[0].Title)
	assert.Equal(t, "self_info", results[0].Engine)
}

func TestSelfInfoPlugin_UserAgent(t *testing.T) {
	p := &selfInfoPlugin{}
	ctx := &plugin.SearchContext{
		Query:       "user-agent",
		Preferences: map[string]any{"user_agent": "Mozilla/5.0"},
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "Your user agent is Mozilla/5.0", results[0].Title)
}

func TestSelfInfoPlugin_NoMatch(t *testing.T) {
	p := &selfInfoPlugin{}
	ctx := &plugin.SearchContext{
		Query:       "something else",
		Preferences: map[string]any{},
	}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestSelfInfoPlugin_NoRemoteAddr(t *testing.T) {
	p := &selfInfoPlugin{}
	ctx := &plugin.SearchContext{
		Query:       "ip",
		Preferences: map[string]any{},
	}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestSelfInfoPlugin_NoUserAgent(t *testing.T) {
	p := &selfInfoPlugin{}
	ctx := &plugin.SearchContext{
		Query:       "user-agent",
		Preferences: map[string]any{},
	}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestSelfInfoPlugin_IPCaseInsensitive(t *testing.T) {
	p := &selfInfoPlugin{}
	ctx := &plugin.SearchContext{
		Query:       "IP",
		Preferences: map[string]any{"remote_addr": "10.0.0.1"},
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "Your IP address is 10.0.0.1", results[0].Title)
}
