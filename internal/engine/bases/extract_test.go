package bases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTMLToText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello World</p>", "Hello World"},
		{"<div>Line 1<br>Line 2</div>", "Line 1\nLine 2"},
		{"<b>Bold</b> and <i>italic</i>", "Bold and italic"},
		{"&amp; &lt; &gt;", "& < >"},
		{"  extra   spaces  ", "extra spaces"},
	}

	for _, tc := range tests {
		got := htmlToText(tc.input)
		assert.Equal(t, tc.expected, got, "input: %s", tc.input)
	}
}

func TestExtractURL(t *testing.T) {
	// Relative URL with base
	got := extractURL("https://example.com/path/", "/search?q=test")
	assert.Equal(t, "https://example.com/search?q=test", got)

	// Already absolute
	got = extractURL("https://example.com", "https://other.com/page")
	assert.Equal(t, "https://other.com/page", got)

	// Empty href
	got = extractURL("https://example.com", "")
	assert.Equal(t, "", got)
}

func TestEvalXPath_GetOne(t *testing.T) {
	doc := mustParseHTML(t, `<html><body><h1>Title</h1><p>Body</p></body></html>`)
	got := evalXPathGetOne(doc, "//h1")
	assert.Equal(t, "Title", got)

	got = evalXPathGetOne(doc, "//nonexistent")
	assert.Equal(t, "", got)
}

func TestEvalXPath_GetAll(t *testing.T) {
	doc := mustParseHTML(t, `<html><body><ul><li>A</li><li>B</li><li>C</li></ul></body></html>`)
	items := evalXPathGetAll(doc, "//li")
	assert.Equal(t, []string{"A", "B", "C"}, items)
}

func TestJSObjStrToJSON(t *testing.T) {
	input := `{title: 'Hello', count: 42, flag: true, items: [1,2,3]}`
	got := jsObjStrToJSON(input)
	assert.Contains(t, got, `"title"`)
	assert.Contains(t, got, `"Hello"`)
	assert.Contains(t, got, `"count"`)
	assert.Contains(t, got, `42`)
}
