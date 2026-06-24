package wikimedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
)

const defaultListURL = "https://meta.wikimedia.org/wiki/List_of_Wikipedias"

var listOfWikipediasURL = defaultListURL

// SparqlEscape escapes a string for use inside a SPARQL string literal.
// It follows the replacement order used by upstream SearXNG.
func SparqlEscape(s string) string {
	repls := []struct{ old, new string }{
		{"\\", "\\\\"},
		{"\t", "\\\t"},
		{"\n", "\\\n"},
		{"\r", "\\\r"},
		{"\b", "\\\b"},
		{"\f", "\\\f"},
		{"\"", "\\\""},
		{"'", "\\'"},
	}
	for _, r := range repls {
		s = strings.ReplaceAll(s, r.old, r.new)
	}
	return s
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// HTMLToText strips HTML tags and normalizes whitespace.
func HTMLToText(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}

// WikiNetlocStore fetches and caches the language -> wikipedia netloc mapping.
type WikiNetlocStore struct {
	mu        sync.RWMutex
	client    *httpx.Client
	cachePath string
}

// NewWikiNetlocStore creates a store. If cachePath is empty it uses the default.
func NewWikiNetlocStore(client *httpx.Client, cachePath string) *WikiNetlocStore {
	if cachePath == "" {
		cachePath = defaultCachePath()
	}
	return &WikiNetlocStore{client: client, cachePath: cachePath}
}

func defaultCachePath() string {
	if p := os.Getenv("SEARGO_WIKI_NETLOC_CACHE"); p != "" {
		return p
	}
	return "data/wiki_netloc.json"
}

// LoadOrFetch returns the mapping. It tries to fetch from Wikimedia; on failure
// it falls back to the local cache file. It returns false only when both fail.
func (s *WikiNetlocStore) LoadOrFetch(ctx context.Context) (map[string]string, bool) {
	cached, ok := s.loadCache()

	fetched, err := s.fetch(ctx)
	if err == nil && len(fetched) > 0 {
		_ = s.saveCache(fetched)
		return fetched, true
	}

	if ok {
		return cached, true
	}
	return nil, false
}

func (s *WikiNetlocStore) fetch(ctx context.Context) (map[string]string, error) {
	if s.client == nil {
		return nil, fmt.Errorf("no http client")
	}
	resp, err := s.client.R().SetContext(ctx).
		SetHeader("Accept", "text/html").
		Get(listOfWikipediasURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]string)
	doc.Find("table.sortable").First().Find("tbody tr").Each(func(_ int, sel *goquery.Selection) {
		link := sel.Find("td").Eq(3).Find("a")
		href, ok := link.Attr("href")
		if !ok {
			return
		}
		engTag := strings.TrimSpace(link.Text())
		if engTag == "" {
			return
		}
		u, err := url.Parse(href)
		if err != nil || u.Host == "" {
			return
		}
		mapping[engTag] = u.Host
	})

	if len(mapping) == 0 {
		return nil, fmt.Errorf("no wikipedia netlocs parsed")
	}
	return mapping, nil
}

func (s *WikiNetlocStore) loadCache() (map[string]string, bool) {
	data, err := os.ReadFile(s.cachePath)
	if err != nil {
		return nil, false
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	return m, true
}

func (s *WikiNetlocStore) saveCache(mapping map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(s.cachePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.cachePath, data, 0o644)
}

// ResolveWikiNetloc selects the language tag and wikipedia netloc for a user
// locale, mirroring upstream get_wiki_params.
func ResolveWikiNetloc(traits engine.EngineTraits, mapping map[string]string, userLocale string) (langTag, netloc string) {
	langTag = ""
	if v, ok := traits.Regions[userLocale]; ok {
		langTag = v
	}
	if langTag == "" {
		if v, ok := traits.Languages[userLocale]; ok {
			langTag = v
		}
	}
	if langTag == "" {
		base := strings.SplitN(userLocale, "-", 2)[0]
		if v, ok := traits.Languages[base]; ok {
			langTag = v
		}
	}
	if langTag == "" {
		langTag = "en"
	}

	netloc = mapping[langTag]
	if netloc == "" {
		netloc = langTag + ".wikipedia.org"
	}
	return langTag, netloc
}
