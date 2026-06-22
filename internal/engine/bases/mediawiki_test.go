package bases

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

func TestMediaWikiEngine_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"query": map[string]interface{}{
				"search": []map[string]interface{}{
					{"title": "Go (programming language)", "pageid": 1, "snippet": "Go is a statically typed..."},
					{"title": "Go (game)", "pageid": 2, "snippet": "Go is a board game..."},
				},
			},
		})
	}))
	defer server.Close()

	eng := NewMediaWikiEngine("test_wiki", []models.Category{models.CategoryGeneral}, MediaWikiConfig{
		BaseURL: server.URL + "/w/api.php",
	})

	ok := eng.Setup(engine.EngineInitConfig{Name: "test_wiki"})
	assert.True(t, ok)

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
	client := httpx.NewClient(reg, "", "test_wiki", "test-ua", 3*time.Second)

	eng.(*mediaWikiEngine).SetClient(client)

	req := &models.Request{Query: "go", Category: models.CategoryGeneral}
	resp, err := eng.Search(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Results, 2)
	assert.Equal(t, "Go (programming language)", resp.Results[0].Title)
	assert.Contains(t, resp.Results[0].URL, "Go_(programming_language)")
}

func TestMediaWikiEngine_InvalidConfig(t *testing.T) {
	eng := NewMediaWikiEngine("bad", nil, MediaWikiConfig{BaseURL: ""})
	ok := eng.Setup(engine.EngineInitConfig{Name: "bad"})
	assert.False(t, ok)
}
