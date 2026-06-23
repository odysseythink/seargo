package wikidata

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func TestMain(m *testing.M) {
	_ = flag.Set("logtostderr", "true")
	os.Exit(m.Run())
}

func TestWikidataEngine_Basic(t *testing.T) {
	w := &Wikidata{}
	ok := w.Init(context.Background(), engine.EngineInitConfig{})
	assert.True(t, ok)
	assert.Equal(t, "wikidata", w.Name())
	assert.Contains(t, w.Categories(), models.CategoryGeneral)
}

func TestWikidataEngine_SearchMockSPARQL(t *testing.T) {
	mockBody := `{
  "results": {
    "bindings": [
      {
        "item": { "type": "uri", "value": "https://www.wikidata.org/entity/Q64" },
        "itemLabel": { "type": "literal", "value": "Berlin" },
        "itemDescription": { "type": "literal", "value": "Capital and largest city of Germany" },
        "P571v": { "type": "literal", "datatype": "http://www.w3.org/2001/XMLSchema#dateTime", "value": "1237-01-01T00:00:00Z" },
        "P17v": { "type": "uri", "value": "https://www.wikidata.org/entity/Q183" },
        "P17l": { "type": "literal", "value": "Germany" },
        "P1082v": { "type": "literal", "value": "3669495" },
        "P856v": { "type": "uri", "value": "https://www.berlin.de" },
        "P625v": { "type": "literal", "value": "Point(13.4050 52.5200)" },
        "article": { "type": "uri", "value": "https://en.wikipedia.org/wiki/Berlin" },
        "P18v": { "type": "literal", "value": "Berlin skyline.jpg" }
      }
    ]
  }
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/sparql-results+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockBody))
	}))
	defer server.Close()

	origEndpoint := sparqlEndpoint
	sparqlEndpoint = server.URL
	defer func() { sparqlEndpoint = origEndpoint }()

	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}
	reg, err := httpx.NewRegistry(cfg)
	require.NoError(t, err)
	defer reg.Close()

	w := &Wikidata{}
	ok := w.Setup(engine.EngineInitConfig{Client: httpx.NewClient(reg, "", "wikidata", "SearGoTest/1.0", 0)})
	require.True(t, ok)

	resp, err := w.Search(context.Background(), &models.Request{
		Query:    "Berlin",
		Category: models.CategoryGeneral,
		Language: "en",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Results, 2)

	var kinds []string
	for _, r := range resp.Results {
		kinds = append(kinds, r.Kind)
	}
	assert.Contains(t, kinds, "infobox")
	assert.Contains(t, kinds, "keyvalue")

	var r models.Result
	for _, ri := range resp.Results {
		if ri.Kind == "infobox" {
			r = ri
			break
		}
	}
	assert.Equal(t, "infobox", r.Kind)
	assert.Equal(t, "Berlin", r.Title)
	assert.Equal(t, "Capital and largest city of Germany", r.Content)
	assert.Equal(t, "wikidata", r.Engine)
	require.NotNil(t, r.Extra)
	// The Wikipedia article URL is preferred as the canonical infobox ID so
	// cross-engine merging can happen.
	assert.Equal(t, "https://en.wikipedia.org/wiki/Berlin", r.Extra["infobox_id"])

	attrs, ok := r.Extra["attributes"].([]results.InfoboxAttribute)
	require.True(t, ok)
	var attrLabels []string
	for _, a := range attrs {
		attrLabels = append(attrLabels, a.Label)
	}
	assert.Contains(t, attrLabels, "inception")
	assert.Contains(t, attrLabels, "country")
	assert.Contains(t, attrLabels, "population")

	urls, ok := r.Extra["urls"].([]results.InfoboxURL)
	require.True(t, ok)
	var urlTitles []string
	for _, u := range urls {
		urlTitles = append(urlTitles, u.Title)
	}
	assert.Contains(t, urlTitles, "Wikidata")
	assert.Contains(t, urlTitles, "Wikipedia")
	assert.Contains(t, urlTitles, "official website")
	assert.Contains(t, urlTitles, "OpenStreetMap")

	assert.Contains(t, r.Extra["img_src"], "commons.wikimedia.org/wiki/Special:FilePath/")
}

func TestWikidataEngine_SearchNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	origEndpoint := sparqlEndpoint
	sparqlEndpoint = server.URL
	defer func() { sparqlEndpoint = origEndpoint }()

	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}
	reg, err := httpx.NewRegistry(cfg)
	require.NoError(t, err)
	defer reg.Close()

	w := &Wikidata{}
	ok := w.Setup(engine.EngineInitConfig{Client: httpx.NewClient(reg, "", "wikidata", "SearGoTest/1.0", 0)})
	require.True(t, ok)

	resp, err := w.Search(context.Background(), &models.Request{
		Query:    "Berlin",
		Category: models.CategoryGeneral,
		Language: "en",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Results)
}
