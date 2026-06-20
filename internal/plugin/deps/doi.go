package deps

import (
	"net/url"
	"regexp"
	"strings"
)

// doiPathRegex matches a DOI in a URL path.
// Format: 10.xxxx/xxxxx (no trailing ?, #, or whitespace).
var doiPathRegex = regexp.MustCompile(`10\.\d{4,9}/[^\s?#]+`)

// ExtractDOI attempts to extract a DOI from a URL.
// It checks the URL path first, then the "doi" query parameter.
// Returns the cleaned DOI and true if a valid DOI (≤50 chars) is found.
func ExtractDOI(rawURL string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}

	// Check path first
	match := doiPathRegex.FindString(u.Path)
	if match != "" {
		doi := cleanDOI(match)
		if len(doi) <= 50 {
			return doi, true
		}
	}

	// Then check "doi" query parameter
	doi := u.Query().Get("doi")
	if doi != "" {
		doi = cleanDOI(doi)
		if len(doi) <= 50 {
			return doi, true
		}
	}

	return "", false
}

// cleanDOI strips trailing punctuation and common file/format suffixes from a DOI.
func cleanDOI(doi string) string {
	suffixes := []string{".pdf", ".html", ".abstract", ".full"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(doi, suffix) {
			doi = strings.TrimSuffix(doi, suffix)
			break
		}
	}
	doi = strings.TrimRight(doi, ".,;:!?")
	return doi
}

// GetDOIResolverURL returns the resolver URL for the given preference.
// It looks up preferred in resolvers; if not found, it tries defaultKey.
// If neither is found, it returns "https://doi.org/" as the fallback.
func GetDOIResolverURL(preferred string, resolvers map[string]string, defaultKey string) string {
	if url, ok := resolvers[preferred]; ok {
		return url
	}
	if url, ok := resolvers[defaultKey]; ok {
		return url
	}
	return "https://doi.org/"
}
