package builtin

import (
	"regexp"
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestHostnamesPlugin_Init_NoConfig(t *testing.T) {
	p := &hostnamesPlugin{}
	ctx := &plugin.AppContext{Config: nil}
	assert.False(t, p.Init(ctx))
	assert.Nil(t, p.config)
}

func TestHostnamesPlugin_Init_WrongType(t *testing.T) {
	p := &hostnamesPlugin{}
	ctx := &plugin.AppContext{Config: "not a config"}
	assert.False(t, p.Init(ctx))
	assert.Nil(t, p.config)
}

func TestHostnamesPlugin_Init_WithConfig(t *testing.T) {
	p := &hostnamesPlugin{}
	cfg := &hostnamesConfig{
		Remove: []*regexp.Regexp{regexp.MustCompile(`example\.com`)},
	}
	ctx := &plugin.AppContext{Config: cfg}
	assert.True(t, p.Init(ctx))
	assert.NotNil(t, p.config)
}

func TestHostnamesPlugin_RemoveMatchingURL(t *testing.T) {
	p := &hostnamesPlugin{
		config: &hostnamesConfig{
			Remove: []*regexp.Regexp{regexp.MustCompile(`spam\.com`)},
		},
	}
	ctx := &plugin.SearchContext{}
	r := &models.Result{URL: "https://spam.com/page"}
	assert.False(t, p.OnResult(ctx, r))
}

func TestHostnamesPlugin_KeepNonMatchingURL(t *testing.T) {
	p := &hostnamesPlugin{
		config: &hostnamesConfig{
			Remove: []*regexp.Regexp{regexp.MustCompile(`spam\.com`)},
		},
	}
	ctx := &plugin.SearchContext{}
	r := &models.Result{URL: "https://good.com/page"}
	assert.True(t, p.OnResult(ctx, r))
	assert.Equal(t, "https://good.com/page", r.URL)
}

func TestHostnamesPlugin_ReplaceURL(t *testing.T) {
	re := regexp.MustCompile(`spam\.com`)
	p := &hostnamesPlugin{
		config: &hostnamesConfig{
			Replace: map[*regexp.Regexp]string{re: "good.com"},
		},
	}
	ctx := &plugin.SearchContext{}
	r := &models.Result{URL: "https://spam.com/page"}
	assert.True(t, p.OnResult(ctx, r))
	assert.Equal(t, "https://good.com/page", r.URL)
}

func TestHostnamesPlugin_NilConfigSafe(t *testing.T) {
	p := &hostnamesPlugin{}
	ctx := &plugin.SearchContext{}
	r := &models.Result{URL: "https://example.com/page"}
	assert.True(t, p.OnResult(ctx, r))
	assert.Equal(t, "https://example.com/page", r.URL)
}
