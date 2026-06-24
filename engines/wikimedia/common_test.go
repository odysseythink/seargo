package wikimedia

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
)

func testClient(t *testing.T) *httpx.Client {
	t.Helper()
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
	}
	reg, err := httpx.NewRegistry(cfg)
	require.NoError(t, err)
	return httpx.NewClient(reg, "", "wikimedia", "SearGoTest/1.0", 0)
}

func TestSparqlEscape(t *testing.T) {
	assert.Equal(t, `\"hello\"`, SparqlEscape(`"hello"`))
	assert.Equal(t, `line1\\line2`, SparqlEscape(`line1\line2`))
	assert.Equal(t, `tab\there`, SparqlEscape("tab\there"))
}

func TestHTMLToText(t *testing.T) {
	assert.Equal(t, "hello world", HTMLToText("<p>hello <b>world</b></p>"))
	assert.Equal(t, "a b", HTMLToText("a\n\n   b  "))
}

func TestResolveWikiNetloc(t *testing.T) {
	traits := engine.EngineTraits{
		Languages: map[string]string{"en": "en", "zh": "zh", "zh_Hans": "zh"},
		Regions:   map[string]string{"zh-CN": "zh", "zh-TW": "zh"},
	}
	mapping := map[string]string{"en": "en.wikipedia.org", "zh": "zh.wikipedia.org"}

	tag, netloc := ResolveWikiNetloc(traits, mapping, "en-US")
	assert.Equal(t, "en", tag)
	assert.Equal(t, "en.wikipedia.org", netloc)

	tag, netloc = ResolveWikiNetloc(traits, mapping, "zh-CN")
	assert.Equal(t, "zh", tag)
	assert.Equal(t, "zh.wikipedia.org", netloc)

	// unknown locale falls back to eng_tag + .wikipedia.org
	tag, netloc = ResolveWikiNetloc(engine.EngineTraits{}, nil, "xx-YY")
	assert.Equal(t, "en", tag)
	assert.Equal(t, "en.wikipedia.org", netloc)
}

func TestWikiNetlocStore_LoadOrFetch_UsesCacheOnFetchFailure(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "wiki_netloc.json")
	cacheData := map[string]string{"en": "en.wikipedia.org"}
	writeCache(t, cachePath, cacheData)

	// Server returns 500, so the store must fall back to the cache file.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	old := listOfWikipediasURL
	listOfWikipediasURL = server.URL
	defer func() { listOfWikipediasURL = old }()

	store := NewWikiNetlocStore(testClient(t), cachePath)
	mapping, ok := store.LoadOrFetch(context.Background())
	require.True(t, ok, "expected cache fallback to succeed")
	assert.Equal(t, "en.wikipedia.org", mapping["en"])
}

func writeCache(t *testing.T, path string, data map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	require.NoError(t, err)
	_, err = fmt.Fprintf(f, "{\n")
	require.NoError(t, err)
	first := true
	for k, v := range data {
		if !first {
			_, err = fmt.Fprintf(f, ",\n")
			require.NoError(t, err)
		}
		_, err = fmt.Fprintf(f, "  %q: %q", k, v)
		require.NoError(t, err)
		first = false
	}
	_, err = fmt.Fprintf(f, "\n}\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())
}
