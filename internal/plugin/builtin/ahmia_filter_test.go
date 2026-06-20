package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/internal/plugin/deps"
	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAhmiaFilter_RemovesBlacklistedOnion(t *testing.T) {
	p := &ahmiaFilterPlugin{
		blacklist:  deps.NewAhmiaBlacklist(),
		torEnabled: true,
	}

	hash := "022bf02537647fc6dadb713f4ee4ff27"
	p.blacklist.Add(hash)

	ctx := &plugin.SearchContext{}
	r := &models.Result{
		URL:     "http://darkbad.onion/evil",
		IsOnion: true,
	}

	ok := p.OnResult(ctx, r)
	assert.False(t, ok, "blacklisted onion result should be removed")
}

func TestAhmiaFilter_KeepsNonOnion(t *testing.T) {
	p := &ahmiaFilterPlugin{
		blacklist:  deps.NewAhmiaBlacklist(),
		torEnabled: true,
	}

	ctx := &plugin.SearchContext{}
	r := &models.Result{
		URL:     "https://example.com/page",
		IsOnion: false,
	}

	ok := p.OnResult(ctx, r)
	assert.True(t, ok, "non-onion result should pass through")
}

func TestAhmiaFilter_KeepsWhitelistedOnion(t *testing.T) {
	p := &ahmiaFilterPlugin{
		blacklist:  deps.NewAhmiaBlacklist(),
		torEnabled: true,
	}

	// Add a different hash so "good.onion" is not blacklisted
	p.blacklist.Add("some_other_hash")

	ctx := &plugin.SearchContext{}
	r := &models.Result{
		URL:     "http://good.onion/page",
		IsOnion: true,
	}

	ok := p.OnResult(ctx, r)
	assert.True(t, ok, "whitelisted onion result should pass through")
}

func TestAhmiaFilter_InitFailsWithoutTor(t *testing.T) {
	p := &ahmiaFilterPlugin{
		blacklist:  deps.NewAhmiaBlacklist(),
		torEnabled: false,
	}

	ok := p.Init(&plugin.AppContext{})
	assert.False(t, ok, "init should return false when Tor is not enabled")
}

func TestAhmiaFilter_InitSucceedsWithTor(t *testing.T) {
	p := &ahmiaFilterPlugin{
		blacklist:  deps.NewAhmiaBlacklist(),
		torEnabled: true,
	}

	ok := p.Init(&plugin.AppContext{})
	assert.True(t, ok, "init should return true when Tor is enabled")
}
