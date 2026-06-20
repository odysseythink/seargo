package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeURL_DefaultScheme(t *testing.T) {
	br := &BaseResult{URL: "example.com/path"}
	normalizeURL(br)
	assert.Equal(t, "http://example.com/path", br.URL)
}

func TestNormalizeURL_HTTPSPreserved(t *testing.T) {
	br := &BaseResult{URL: "https://example.com/path"}
	original := br.URL
	normalizeURL(br)
	assert.Equal(t, original, br.URL)
}

func TestNormalizeURL_SyncParsedURL(t *testing.T) {
	br := &BaseResult{URL: "https://example.com/path?q=1#frag"}
	normalizeURL(br)
	assert.Equal(t, []string{"https", "example.com", "/path", "q=1", "frag"}, br.ParsedURL)
}

func TestNormalizeURL_Invalid(t *testing.T) {
	br := &BaseResult{URL: "://invalid"}
	normalizeURL(br)
	assert.Equal(t, "://invalid", br.URL)
}

func TestNormalizeText_CollapseWhitespace(t *testing.T) {
	br := &BaseResult{Title: "  Hello   World  ", Content: "Line1\n\nLine2"}
	normalizeText(br)
	assert.Equal(t, "Hello World", br.Title)
	assert.Equal(t, "Line1 Line2", br.Content)
}

func TestNormalizeText_DedupTitleEqualsContent(t *testing.T) {
	br := &BaseResult{Title: "Same text", Content: "Same text"}
	normalizeText(br)
	assert.Equal(t, "", br.Content, "content should be cleared when identical to title")
}

func TestNormalizeText_TitleDiffersFromContent(t *testing.T) {
	br := &BaseResult{Title: "Title", Content: "Different content"}
	normalizeText(br)
	assert.Equal(t, "Title", br.Title)
	assert.Equal(t, "Different content", br.Content)
}

func TestNormalizeDate_ValidYear(t *testing.T) {
	br := &BaseResult{Title: "T"}
	normalizeDate(br, "2024")
	assert.NotNil(t, br.PublishedAt)
	assert.Equal(t, 2024, br.PublishedAt.Year())
}

func TestNormalizeDate_YearTooLow(t *testing.T) {
	br := &BaseResult{Title: "T"}
	normalizeDate(br, "1899")
	assert.Nil(t, br.PublishedAt, "year < 1900 should be rejected")
}

func TestNormalizeDate_InvalidFormat(t *testing.T) {
	br := &BaseResult{Title: "T"}
	normalizeDate(br, "not-a-date")
	assert.Nil(t, br.PublishedAt)
}
