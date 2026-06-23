package wikicommons

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
	"github.com/seargo/seargo/pkg/models/results"
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
	return httpx.NewClient(reg, "", "wikicommons", "test-ua", 0)
}

func TestWikicommons_SearchFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("gsrsearch"), "filetype:multimedia")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"query": map[string]interface{}{
				"pages": map[string]interface{}{
					"1": map[string]interface{}{
						"title":   "File:Example.pdf",
						"snippet": "example",
						"imageinfo": []map[string]interface{}{
							{
								"descriptionurl": "https://commons.wikimedia.org/wiki/File:Example.pdf",
								"url":            "https://upload.wikimedia.org/Example.pdf",
								"mime":           "application/pdf",
								"thumburl":       "https://thumb.png",
								"size":           2048,
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	old := apiURL
	apiURL = server.URL
	defer func() { apiURL = old }()

	engine.Reset()
	engine.Register("wikicommons", &Wikicommons{})
	eng, _ := engine.Get("wikicommons")

	require.True(t, eng.Setup(engine.EngineInitConfig{
		Client:     testClient(t),
		Categories: []models.Category{models.CategoryFiles},
		Extra:      map[string]any{"wc_search_type": "file"},
	}))

	resp, err := eng.Search(context.Background(), &models.Request{
		Query:    "pdf",
		Category: models.CategoryFiles,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "file", resp.Results[0].Kind)
	assert.Equal(t, "files.html", resp.Results[0].Template)
	require.Len(t, resp.TypedResults, 1)
	fr, ok := resp.TypedResults[0].(*results.FileResult)
	require.True(t, ok)
	assert.Equal(t, "Example", fr.Title)
	assert.Equal(t, "application/pdf", fr.FileType)
	assert.Equal(t, int64(2048), fr.FileSize)
}
