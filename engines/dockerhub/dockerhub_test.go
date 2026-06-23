package dockerhub

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
	return httpx.NewClient(reg, "", "docker_hub", "test-ua", 0)
}

func TestDockerHub_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "golang", r.URL.Query().Get("query"))
		assert.Equal(t, "0", r.URL.Query().Get("from"))
		assert.Equal(t, "10", r.URL.Query().Get("size"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"name":               "golang",
					"slug":               "library/golang",
					"source":             "official",
					"short_description":  "Go compiler",
					"logo_url":           map[string]string{"large": "https://hub.docker.com/logo.png"},
				},
			},
		})
	}))
	defer server.Close()

	old := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = old }()

	eng := &DockerHub{}
	require.True(t, eng.Setup(engine.EngineInitConfig{Client: testClient(t)}))

	resp, err := eng.Search(context.Background(), &models.Request{Query: "golang", Category: models.CategoryIT})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "golang", resp.Results[0].Title)
	assert.Contains(t, resp.Results[0].URL, "/_/library/golang")
	assert.Equal(t, "Go compiler", resp.Results[0].Content)
}
