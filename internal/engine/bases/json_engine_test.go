package bases

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
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/pkg/models"
)

func TestJSONEngine_Search(t *testing.T) {
	logger.Init("error", "stderr")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"response": map[string]interface{}{
				"docs": []map[string]interface{}{
					{"title": "JSON Result 1", "url": "https://example.com/1", "snippet": "First match"},
					{"title": "JSON Result 2", "url": "https://example.com/2", "snippet": "Second match"},
				},
			},
		})
	}))
	defer server.Close()

	eng := NewJSONEngine("test_json", []models.Category{models.CategoryGeneral}, JSONEngineConfig{
		SearchURL:    server.URL + "/api?q={query}",
		ResultsQuery: "response/docs",
		URLQuery:     "url",
		TitleQuery:   "title",
		ContentQuery: "snippet",
	})

	ok := eng.Setup(engine.EngineInitConfig{Name: "test_json"})
	assert.True(t, ok)

	reg, err := httpx.NewRegistry(&config.Config{
		Outgoing: config.OutgoingConfig{
			EnableHTTP:     true,
			RequestTimeout: 10.0,
			MaxRedirects:   5,
			EnableHTTP2:    true,
		},
	})
	require.NoError(t, err)

	client := httpx.NewClient(reg, "", "test_json", "test-ua", 0)
	eng.(*jsonEngine).SetClient(client)

	req := &models.Request{Query: "test", Category: models.CategoryGeneral}
	resp, err := eng.Search(context.Background(), req)
	require.NoError(t, err)
	assert.Len(t, resp.Results, 2)
	assert.Equal(t, "JSON Result 1", resp.Results[0].Title)
}

func TestJSONEngine_InvalidConfig(t *testing.T) {
	eng := NewJSONEngine("bad", nil, JSONEngineConfig{
		SearchURL: "",
	})
	ok := eng.Setup(engine.EngineInitConfig{Name: "bad"})
	assert.False(t, ok, "engine without search URL should fail Setup")
}
