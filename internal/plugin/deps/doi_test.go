package deps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDOIPath(t *testing.T) {
	doi, ok := ExtractDOI("https://doi.org/10.1234/abc123")
	assert.True(t, ok)
	assert.Equal(t, "10.1234/abc123", doi)
}

func TestExtractDOIPathWithSubpath(t *testing.T) {
	doi, ok := ExtractDOI("https://example.com/articles/10.5678/def456")
	assert.True(t, ok)
	assert.Equal(t, "10.5678/def456", doi)
}

func TestExtractDOIParam(t *testing.T) {
	doi, ok := ExtractDOI("https://example.com/article?doi=10.1234/abc123")
	assert.True(t, ok)
	assert.Equal(t, "10.1234/abc123", doi)
}

func TestNoDOI(t *testing.T) {
	_, ok := ExtractDOI("https://example.com/page")
	assert.False(t, ok)
}

func TestLongDOIRejected(t *testing.T) {
	// DOI longer than 50 characters should be rejected
	_, ok := ExtractDOI("https://doi.org/10.1234/abcdefghijklmnopqrstuvwxyz1234567890abc1234567890")
	assert.False(t, ok)
}

func TestCleanDOIPdfSuffix(t *testing.T) {
	doi, ok := ExtractDOI("https://doi.org/10.1234/abc.pdf")
	assert.True(t, ok)
	assert.Equal(t, "10.1234/abc", doi)
}

func TestCleanDOIHtmlSuffix(t *testing.T) {
	doi, ok := ExtractDOI("https://doi.org/10.1234/abc.html")
	assert.True(t, ok)
	assert.Equal(t, "10.1234/abc", doi)
}

func TestResolverFound(t *testing.T) {
	resolvers := map[string]string{
		"google": "https://scholar.google.com/",
	}
	url := GetDOIResolverURL("google", resolvers, "default")
	assert.Equal(t, "https://scholar.google.com/", url)
}

func TestResolverFallbackToDefaultKey(t *testing.T) {
	resolvers := map[string]string{
		"default": "https://sci-hub.se/",
	}
	url := GetDOIResolverURL("unknown", resolvers, "default")
	assert.Equal(t, "https://sci-hub.se/", url)
}

func TestResolverFallbackToDoiOrg(t *testing.T) {
	resolvers := map[string]string{}
	url := GetDOIResolverURL("unknown", resolvers, "default")
	assert.Equal(t, "https://doi.org/", url)
}
