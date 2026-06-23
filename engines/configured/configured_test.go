package configured

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
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
	return httpx.NewClient(reg, "", "mdn", "test-ua", 0)
}

func TestConfigured_JSON_URLPrefixAndPaging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "array", r.URL.Query().Get("q"))
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"documents": []map[string]interface{}{
				{
					"title":   "Array",
					"summary": "JS arrays",
					"mdn_url": "/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array",
				},
			},
		})
	}))
	defer server.Close()

	engine.Reset()
	eng := &configuredEngine{name: "mdn", defaultCategories: []models.Category{models.CategoryIT}}

	cfg := engine.EngineInitConfig{
		Client:     testClient(t),
		Categories: []models.Category{models.CategoryIT},
		Extra: map[string]any{
			"search_url":    server.URL + "/search?q={query}&page={page}",
			"results_query": "documents",
			"url_query":     "mdn_url",
			"title_query":   "title",
			"content_query": "summary",
			"url_prefix":    "https://developer.mozilla.org",
			"paging":        true,
		},
	}
	require.True(t, eng.Setup(cfg))

	resp, err := eng.Search(context.Background(), &models.Request{
		Query:    "array",
		Category: models.CategoryIT,
		Page:     2,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "Array", resp.Results[0].Title)
	assert.Equal(t, "https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array", resp.Results[0].URL)
}
