package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestTrackerURLRemover_StripsUTMParams(t *testing.T) {
	p := &trackerURLRemoverPlugin{}
	ctx := &plugin.SearchContext{}

	// Non-onion result — should pass through even with IsOnion=false
	r := &models.Result{
		URL:          "https://example.com/page?utm_source=google&utm_medium=cpc&foo=bar",
		ThumbnailURL: "https://cdn.example.com/thumb.jpg?utm_campaign=summer",
	}

	ok := p.OnResult(ctx, r)
	assert.True(t, ok)

	assert.NotContains(t, r.URL, "utm_source")
	assert.NotContains(t, r.URL, "utm_medium")
	assert.Contains(t, r.URL, "foo=bar")

	assert.NotContains(t, r.ThumbnailURL, "utm_campaign")
}

func TestTrackerURLRemover_URLWithoutTrackingUnchanged(t *testing.T) {
	p := &trackerURLRemoverPlugin{}
	ctx := &plugin.SearchContext{}

	original := "https://example.com/page?foo=bar&baz=qux"
	r := &models.Result{URL: original}

	ok := p.OnResult(ctx, r)
	assert.True(t, ok)
	assert.Equal(t, original, r.URL)
}

func TestTrackerURLRemover_EmptyURL(t *testing.T) {
	p := &trackerURLRemoverPlugin{}
	ctx := &plugin.SearchContext{}

	r := &models.Result{Title: "empty URL result"}

	ok := p.OnResult(ctx, r)
	assert.True(t, ok)
	assert.Equal(t, "", r.URL)
}
