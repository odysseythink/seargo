package bases

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// htmlToText strips HTML tags and decodes entities, producing plain text.
// <br> and block-level elements become newlines. Whitespace is collapsed.
func htmlToText(htmlStr string) string {
	const sentinel = "\x00"

	// Remove scripts and styles
	htmlStr = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`).ReplaceAllString(htmlStr, "")
	htmlStr = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`).ReplaceAllString(htmlStr, "")

	// Replace <br> and block elements with sentinel (preserve through collapse)
	blockTags := regexp.MustCompile(`(?i)</?(?:br|p|div|h[1-6]|li|tr|article|section)[^>]*/?>`)
	htmlStr = blockTags.ReplaceAllString(htmlStr, sentinel)

	// Strip remaining tags
	htmlStr = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlStr, "")

	// Decode common entities
	htmlStr = strings.ReplaceAll(htmlStr, "&amp;", "&")
	htmlStr = strings.ReplaceAll(htmlStr, "&lt;", "<")
	htmlStr = strings.ReplaceAll(htmlStr, "&gt;", ">")
	htmlStr = strings.ReplaceAll(htmlStr, "&quot;", "\"")
	htmlStr = strings.ReplaceAll(htmlStr, "&#39;", "'")
	htmlStr = strings.ReplaceAll(htmlStr, "&apos;", "'")
	htmlStr = regexp.MustCompile(`&#x?[0-9a-fA-F]+;`).ReplaceAllString(htmlStr, "")

	// Collapse whitespace (sentinel survives this)
	htmlStr = regexp.MustCompile(`[ \t]+`).ReplaceAllString(htmlStr, " ")
	htmlStr = regexp.MustCompile(`\n+`).ReplaceAllString(htmlStr, "\n")
	// Trim whitespace (but not sentinel yet)
	htmlStr = strings.TrimSpace(htmlStr)

	// Convert sentinels to newlines after stripping surrounding spaces
	htmlStr = regexp.MustCompile(` ?`+regexp.QuoteMeta(sentinel)+` ?`).ReplaceAllString(htmlStr, "\n")
	htmlStr = strings.TrimSpace(htmlStr)

	return htmlStr
}

// extractURL resolves a potentially relative URL against a base URL.
// Returns empty string if href is empty.
func extractURL(baseURL, href string) string {
	if href == "" {
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}

	ref, err := url.Parse(href)
	if err != nil {
		return href
	}

	resolved := base.ResolveReference(ref)
	return resolved.String()
}

// evalXPathGetOne evaluates an XPath expression against an HTML node
// and returns the text content of the first matching element.
func evalXPathGetOne(doc *html.Node, xpath string) string {
	node, err := htmlquery.Query(doc, xpath)
	if err != nil || node == nil {
		return ""
	}
	return strings.TrimSpace(htmlquery.InnerText(node))
}

// evalXPathGetAll evaluates an XPath expression and returns text content
// of all matching elements.
func evalXPathGetAll(doc *html.Node, xpath string) []string {
	nodes, err := htmlquery.QueryAll(doc, xpath)
	if err != nil {
		return nil
	}
	result := make([]string, 0, len(nodes))
	for _, node := range nodes {
		text := strings.TrimSpace(htmlquery.InnerText(node))
		if text != "" {
			result = append(result, text)
		}
	}
	return result
}

// cssGetOne evaluates a CSS selector against a goquery document and returns
// the text content of the first matching element.
func cssGetOne(doc *goquery.Document, selector string) string {
	sel := doc.Find(selector)
	if sel.Length() == 0 {
		return ""
	}
	return strings.TrimSpace(sel.First().Text())
}

// cssGetAll evaluates a CSS selector and returns text content of all matches.
func cssGetAll(doc *goquery.Document, selector string) []string {
	var result []string
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			result = append(result, text)
		}
	})
	return result
}

// cssGetAttr returns an attribute value from the first element matching the
// CSS selector.
func cssGetAttr(doc *goquery.Document, selector, attr string) string {
	sel := doc.Find(selector)
	if sel.Length() == 0 {
		return ""
	}
	val, _ := sel.First().Attr(attr)
	return strings.TrimSpace(val)
}

// jsObjStrToJSON converts a JavaScript object literal string to valid JSON
// by quoting unquoted keys. Handles single-quoted strings.
func jsObjStrToJSON(s string) string {
	// Quote unquoted keys: {key: value} → {"key": value}
	re := regexp.MustCompile(`([{,])\s*([a-zA-Z_$][a-zA-Z0-9_$]*)\s*:`)
	s = re.ReplaceAllString(s, `${1}"$2":`)

	// Replace single-quoted strings with double-quoted
	s = regexp.MustCompile(`'([^'\\]*(\\.[^'\\]*)*)'`).ReplaceAllString(s, `"$1"`)

	return s
}

// mustParseHTML is a test helper that parses HTML and panics on error.
func mustParseHTML(t interface{ Fatalf(string, ...interface{}) }, htmlStr string) *html.Node {
	doc, err := htmlquery.Parse(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}
	return doc
}
