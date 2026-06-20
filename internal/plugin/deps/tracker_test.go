package deps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveUTMParams(t *testing.T) {
	original := "https://example.com/page?utm_source=google&utm_medium=cpc&foo=bar"
	cleaned, changed := TrackerCleanURL(original)
	assert.True(t, changed)
	assert.NotContains(t, cleaned, "utm_source")
	assert.NotContains(t, cleaned, "utm_medium")
	assert.Contains(t, cleaned, "foo=bar")
}

func TestRemoveFBCLID(t *testing.T) {
	original := "https://example.com/page?fbclid=abc123&q=search"
	cleaned, changed := TrackerCleanURL(original)
	assert.True(t, changed)
	assert.NotContains(t, cleaned, "fbclid")
	assert.Contains(t, cleaned, "q=search")
}

func TestRemoveAllTrackingParams(t *testing.T) {
	original := "https://example.com/?utm_source=google&utm_medium=cpc&utm_campaign=summer&utm_term=shoes&utm_content=ad1&fbclid=abc&gclid=def&_ga=ghi"
	cleaned, changed := TrackerCleanURL(original)
	assert.True(t, changed)
	assert.NotContains(t, cleaned, "utm_")
	assert.NotContains(t, cleaned, "fbclid")
	assert.NotContains(t, cleaned, "gclid")
	assert.NotContains(t, cleaned, "_ga")
}

func TestURLWithoutTracking(t *testing.T) {
	original := "https://example.com/page?foo=bar&baz=qux"
	cleaned, changed := TrackerCleanURL(original)
	assert.False(t, changed)
	assert.Equal(t, original, cleaned)
}

func TestEmptyURL(t *testing.T) {
	cleaned, changed := TrackerCleanURL("")
	assert.False(t, changed)
	assert.Equal(t, "", cleaned)
}

func TestURLWithFragmentsPreserved(t *testing.T) {
	original := "https://example.com/page?utm_source=google&section=1#fragment"
	cleaned, changed := TrackerCleanURL(original)
	assert.True(t, changed)
	assert.NotContains(t, cleaned, "utm_source")
	assert.Contains(t, cleaned, "#fragment")
}
