package wikidata

import (
	"context"
	"encoding/json"
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
)

func TestMain(m *testing.M) {
	_ = flag.Set("logtostderr", "true")
	os.Exit(m.Run())
}

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
	return httpx.NewClient(reg, "", "wikidata", "SearGoTest/1.0", 0)
}

func TestWikidataEngine_LimitOneNoKeyValue(t *testing.T) {
	mockBody := map[string]any{
		"results": map[string]any{
			"bindings": []map[string]any{
				{
					"item":            map[string]any{"type": "uri", "value": "https://www.wikidata.org/entity/Q64"},
					"itemLabel":       map[string]any{"type": "literal", "value": "Berlin"},
					"itemDescription": map[string]any{"type": "literal", "value": "Capital of Germany"},
					"P18s":            map[string]any{"type": "literal", "value": "https://commons.wikimedia.org/wiki/Special:FilePath/Berlin%20skyline.jpg?width=300"},
					"articleen":       map[string]any{"type": "uri", "value": "https://en.wikipedia.org/wiki/Berlin"},
				},
				{
					"item":            map[string]any{"type": "uri", "value": "https://www.wikidata.org/entity/Q12345"},
					"itemLabel":       map[string]any{"type": "literal", "value": "Other"},
					"itemDescription": map[string]any{"type": "literal", "value": "Other entity"},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/sparql-results+json")
		json.NewEncoder(w).Encode(mockBody)
	}))
	defer server.Close()

	orig := sparqlEndpoint
	sparqlEndpoint = server.URL
	defer func() { sparqlEndpoint = orig }()

	w := &Wikidata{}
	require.True(t, w.Setup(engine.EngineInitConfig{Client: testClient(t)}))

	resp, err := w.Search(context.Background(), &models.Request{
		Query:    "Berlin",
		Category: models.CategoryGeneral,
		Language: "en",
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "infobox", resp.Results[0].Kind)
	assert.Equal(t, "Berlin", resp.Results[0].Title)
	assert.Equal(t, "wikidata", resp.Results[0].Engine)

	for _, r := range resp.Results {
		assert.NotEqual(t, "keyvalue", r.Kind)
	}
}
