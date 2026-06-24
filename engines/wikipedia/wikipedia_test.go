package wikipedia

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
	return httpx.NewClient(reg, "", "wikipedia", "SearGoTest/1.0", 0)
}

func TestWikipediaEngine_StandardReturnsInfobox(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "Berlin")
		json.NewEncoder(w).Encode(map[string]any{
			"type":        "standard",
			"title":       "Berlin",
			"titles":      map[string]any{"display": "Berlin"},
			"extract":     "Capital of Germany",
			"description": "Capital and largest city of Germany",
			"content_urls": map[string]any{
				"desktop": map[string]any{"page": "https://en.wikipedia.org/wiki/Berlin"},
			},
			"thumbnail": map[string]any{"source": "https://upload.wikimedia.org/wikipedia/commons/thumb/x/xy/Berlin.jpg/300px-Berlin.jpg"},
		})
	}))
	defer server.Close()

	orig := restSummaryURL
	restSummaryURL = server.URL + "/%s/%s"
	defer func() { restSummaryURL = orig }()

	w := &Wikipedia{}
	require.True(t, w.Setup(engine.EngineInitConfig{Client: testClient(t)}))

	resp, err := w.Search(context.Background(), &models.Request{
		Query:    "berlin",
		Category: models.CategoryGeneral,
		Language: "en",
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "infobox", resp.Results[0].Kind)
	assert.Equal(t, "Berlin", resp.Results[0].Title)
	assert.Equal(t, "Capital of Germany", resp.Results[0].Content)
	assert.Equal(t, "wikipedia", resp.Results[0].Engine)
	require.NotNil(t, resp.Results[0].Extra)
	assert.Equal(t, "https://en.wikipedia.org/wiki/Berlin", resp.Results[0].Extra["infobox_id"])
}

func TestWikipediaEngine_NonStandardReturnsMainResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"type":        "disambiguation",
			"title":       "Golang",
			"titles":      map[string]any{"display": "Golang"},
			"description": "Go programming language",
			"content_urls": map[string]any{
				"desktop": map[string]any{"page": "https://en.wikipedia.org/wiki/Golang"},
			},
		})
	}))
	defer server.Close()

	orig := restSummaryURL
	restSummaryURL = server.URL + "/%s/%s"
	defer func() { restSummaryURL = orig }()

	w := &Wikipedia{}
	require.True(t, w.Setup(engine.EngineInitConfig{Client: testClient(t)}))

	resp, err := w.Search(context.Background(), &models.Request{
		Query:    "golang",
		Category: models.CategoryGeneral,
		Language: "en",
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "main", resp.Results[0].Kind)
	assert.Equal(t, "Golang", resp.Results[0].Title)
	assert.Equal(t, "wikipedia", resp.Results[0].Engine)
}

func TestWikipediaEngine_404ReturnsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	orig := restSummaryURL
	restSummaryURL = server.URL + "/%s/%s"
	defer func() { restSummaryURL = orig }()

	w := &Wikipedia{}
	require.True(t, w.Setup(engine.EngineInitConfig{Client: testClient(t)}))

	resp, err := w.Search(context.Background(), &models.Request{
		Query:    "xyznonexistent",
		Category: models.CategoryGeneral,
		Language: "en",
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Results)
}
