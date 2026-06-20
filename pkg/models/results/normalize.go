package results

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var whitespaceRE = regexp.MustCompile(`\s+`)

// normalizeURL normalizes a result URL:
// - Default scheme to http if missing
// - Sync ParsedURL with parsed components
func normalizeURL(r *BaseResult) {
	if r.URL == "" {
		return
	}

	raw := r.URL

	// Default scheme
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		r.URL = raw
		r.ParsedURL = nil
		return
	}

	r.URL = u.String()
	r.ParsedURL = []string{u.Scheme, u.Host, u.Path, u.RawQuery, u.Fragment}
}

// normalizeText collapses whitespace and deduplicates title==content.
func normalizeText(r *BaseResult) {
	r.Title = strings.TrimSpace(whitespaceRE.ReplaceAllString(r.Title, " "))
	r.Content = strings.TrimSpace(whitespaceRE.ReplaceAllString(r.Content, " "))

	// Dedup: if content equals title, clear content
	if r.Content != "" && r.Content == r.Title {
		r.Content = ""
	}
}

// normalizeEngines ensures Engine is present in Engines.
func normalizeEngines(r *BaseResult) {
	if r.Engine != "" {
		for _, e := range r.Engines {
			if e == r.Engine {
				return
			}
		}
		r.Engines = append(r.Engines, r.Engine)
	}
}

// normalizeDate attempts to parse a date string and set PublishedAt.
// Years < 1900 are rejected as invalid.
func normalizeDate(r *BaseResult, dateStr string) {
	if dateStr == "" {
		return
	}

	// Try common formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
		"2006-01",
		"2006",
		"02 Jan 2006",
		"January 2, 2006",
		"Jan 2, 2006",
	}

	for _, layout := range formats {
		t, err := time.Parse(layout, dateStr)
		if err == nil {
			if t.Year() >= 1900 {
				r.PublishedAt = &t
				return
			}
			return
		}
	}

	// Try parsing as year only
	if year, err := strconv.Atoi(dateStr); err == nil && year >= 1900 {
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		r.PublishedAt = &t
	}
}

// Normalize applies all normalization functions to a BaseResult.
func (r *BaseResult) Normalize() {
	normalizeURL(r)
	normalizeText(r)
	normalizeEngines(r)
}

// PostNormalize is a default no-op; concrete types override for kind-specific logic.
func (r *BaseResult) PostNormalize() {}

// NormalizeResult applies base normalization via the Result interface.
func NormalizeResult(r Result) {
	r.Base().Normalize()
}

// extractDomainFromURL extracts the domain (host) from a URL string.
func extractDomainFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

// computeParsedURL fills ParsedURL from the URL string.
func computeParsedURL(rawURL string) []string {
	if rawURL == "" {
		return nil
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	return []string{u.Scheme, u.Host, u.Path, u.RawQuery, u.Fragment}
}
