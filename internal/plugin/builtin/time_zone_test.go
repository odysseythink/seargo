package builtin

import (
	"strings"
	"testing"
	"time"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/stretchr/testify/assert"
)

func TestTimeZonePlugin_SpecificCity(t *testing.T) {
	p := &timeZonePlugin{}
	ctx := &plugin.SearchContext{Query: "time Berlin"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "Current time in Berlin", results[0].Title)
	assert.Contains(t, results[0].Content, "Europe/Berlin")
	assert.Equal(t, "time_zone", results[0].Engine)
}

func TestTimeZonePlugin_CurrentTime(t *testing.T) {
	p := &timeZonePlugin{}
	ctx := &plugin.SearchContext{Query: "time"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "Current time in UTC", results[0].Title)
	assert.Contains(t, results[0].Content, "UTC")
}

func TestTimeZonePlugin_UnknownCity(t *testing.T) {
	p := &timeZonePlugin{}
	ctx := &plugin.SearchContext{Query: "time Atlantis"}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestTimeZonePlugin_NoMatch(t *testing.T) {
	p := &timeZonePlugin{}
	ctx := &plugin.SearchContext{Query: "weather Berlin"}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestTimeZonePlugin_TimezoneKeyword(t *testing.T) {
	p := &timeZonePlugin{}
	ctx := &plugin.SearchContext{Query: "timezone Tokyo"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "Current time in Tokyo", results[0].Title)
	assert.Contains(t, results[0].Content, "Asia/Tokyo")
}

func TestTimeZonePlugin_NowKeyword(t *testing.T) {
	p := &timeZonePlugin{}
	ctx := &plugin.SearchContext{Query: "now London"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "Current time in London", results[0].Title)
	assert.Contains(t, results[0].Content, "Europe/London")
}

func TestTimeZonePlugin_ContentFormat(t *testing.T) {
	p := &timeZonePlugin{}
	ctx := &plugin.SearchContext{Query: "time Berlin"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)

	// Content should match format like "2026-06-20 17:57 Europe/Berlin"
	// Split into at most 3 parts: date, time, timezone
	parts := strings.SplitN(results[0].Content, " ", 3)
	assert.Len(t, parts, 3, "content should have date, time, and timezone")

	// First part should be a date string
	_, err := time.Parse("2006-01-02", parts[0])
	assert.NoError(t, err, "first part should be a valid date")

	// Second part should be a time string
	_, err = time.Parse("15:04", parts[1])
	assert.NoError(t, err, "second part should be a valid time")

	// Third part should be a timezone identifier
	assert.Equal(t, "Europe/Berlin", parts[2])
}
