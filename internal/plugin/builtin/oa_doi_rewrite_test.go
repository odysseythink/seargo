package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestOADOIRewrite_DOIURLRewritten(t *testing.T) {
	p := &oaDOIRewritePlugin{
		preferredResolver: "scihub",
		resolvers: map[string]string{
			"scihub": "https://sci-hub.se/",
		},
		defaultResolver: "scihub",
	}

	ctx := &plugin.SearchContext{}
	r := &models.Result{
		URL:          "https://doi.org/10.1234/abc123",
		ThumbnailURL: "https://cdn.example.com/thumb.jpg",
	}

	ok := p.OnResult(ctx, r)
	assert.True(t, ok)
	assert.Equal(t, "https://sci-hub.se/10.1234/abc123", r.URL)
	// Thumbnail and favicon should remain unchanged
	assert.Equal(t, "https://cdn.example.com/thumb.jpg", r.ThumbnailURL)
}

func TestOADOIRewrite_NoDOIUnchanged(t *testing.T) {
	p := &oaDOIRewritePlugin{
		preferredResolver: "scihub",
		resolvers: map[string]string{
			"scihub": "https://sci-hub.se/",
		},
		defaultResolver: "scihub",
	}

	ctx := &plugin.SearchContext{}
	r := &models.Result{
		URL: "https://example.com/page?foo=bar",
	}

	ok := p.OnResult(ctx, r)
	assert.True(t, ok)
	assert.Equal(t, "https://example.com/page?foo=bar", r.URL)
}

func TestOADOIRewrite_OnlyMainURLRewritten(t *testing.T) {
	p := &oaDOIRewritePlugin{
		preferredResolver: "google",
		resolvers: map[string]string{
			"google":  "https://scholar.google.com/",
			"default": "https://doi.org/",
		},
		defaultResolver: "default",
	}

	ctx := &plugin.SearchContext{}
	r := &models.Result{
		URL:          "https://doi.org/10.5678/def456",
		ThumbnailURL: "https://cdn.example.com/thumb.jpg",
		Favicon:      "https://cdn.example.com/favicon.ico",
	}

	ok := p.OnResult(ctx, r)
	assert.True(t, ok)
	assert.Equal(t, "https://scholar.google.com/10.5678/def456", r.URL)
	assert.Equal(t, "https://cdn.example.com/thumb.jpg", r.ThumbnailURL)
	assert.Equal(t, "https://cdn.example.com/favicon.ico", r.Favicon)
}

func TestOADOIRewrite_InitFromConfig(t *testing.T) {
	cfg := &config.Config{
		DefaultDOIResolver: "oadoi.org",
		DOIRsolvers: map[string]string{
			"oadoi.org": "https://oadoi.org/",
		},
	}

	p := &oaDOIRewritePlugin{}
	ok := p.Init(&plugin.AppContext{Config: cfg})
	assert.True(t, ok)

	r := &models.Result{URL: "https://doi.org/10.1234/abc123"}
	ok = p.OnResult(&plugin.SearchContext{}, r)
	assert.True(t, ok)
	assert.Equal(t, "https://oadoi.org/10.1234/abc123", r.URL)
}

func TestOADOIRewrite_InitUsesDefaultsWhenConfigEmpty(t *testing.T) {
	cfg := &config.Config{}

	p := &oaDOIRewritePlugin{}
	ok := p.Init(&plugin.AppContext{Config: cfg})
	assert.True(t, ok)

	r := &models.Result{URL: "https://doi.org/10.1234/abc123"}
	ok = p.OnResult(&plugin.SearchContext{}, r)
	assert.True(t, ok)
	assert.Equal(t, "https://oadoi.org/10.1234/abc123", r.URL)
}

func TestOADOIRewrite_InitWithInvalidResolverFallsBack(t *testing.T) {
	cfg := &config.Config{
		DefaultDOIResolver: "unknown",
		DOIRsolvers: map[string]string{
			"doi.org": "https://doi.org/",
		},
	}

	p := &oaDOIRewritePlugin{}
	ok := p.Init(&plugin.AppContext{Config: cfg})
	assert.True(t, ok)

	r := &models.Result{URL: "https://doi.org/10.1234/abc123"}
	ok = p.OnResult(&plugin.SearchContext{}, r)
	assert.True(t, ok)
	assert.Equal(t, "https://doi.org/10.1234/abc123", r.URL)
}

func TestOADOIRewrite_InitWithNilConfig(t *testing.T) {
	p := &oaDOIRewritePlugin{}
	ok := p.Init(&plugin.AppContext{Config: nil})
	assert.True(t, ok)

	r := &models.Result{URL: "https://doi.org/10.1234/abc123"}
	ok = p.OnResult(&plugin.SearchContext{}, r)
	assert.True(t, ok)
	assert.Equal(t, "https://oadoi.org/10.1234/abc123", r.URL)
}
