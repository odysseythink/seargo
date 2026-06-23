package stackexchange

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	return httpx.NewClient(reg, "", "stackexchange", "test-ua", 0)
}

func TestStackExchange_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "golang", r.URL.Query().Get("q"))
		assert.Equal(t, "stackoverflow", r.URL.Query().Get("site"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"question_id":   1234,
					"title":         "How do I write Go?",
					"tags":          []string{"go", "golang"},
					"score":         42,
					"is_answered":   true,
					"owner":         map[string]string{"display_name": "alice"},
				},
			},
		})
	}))
	defer server.Close()

	old := searchAPI
	searchAPI = server.URL + "?"
	defer func() { searchAPI = old }()

	engine.Reset()
	engine.Register("stackexchange", &StackExchange{})
	eng, _ := engine.Get("stackexchange")

	require.True(t, eng.Setup(engine.EngineInitConfig{
		Client:     testClient(t),
		Categories: []models.Category{models.CategoryIT},
		Extra:      map[string]any{"api_site": "stackoverflow"},
	}))

	resp, err := eng.Search(context.Background(), &models.Request{
		Query:    "golang",
		Category: models.CategoryIT,
		Page:     1,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "How do I write Go?", resp.Results[0].Title)
	assert.True(t, strings.Contains(resp.Results[0].URL, "stackoverflow.com/q/1234"))
	assert.Contains(t, resp.Results[0].Content, "go, golang")
	assert.Contains(t, resp.Results[0].Content, "score: 42")
}
