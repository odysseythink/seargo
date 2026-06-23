package githubcode

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
	return httpx.NewClient(reg, "", "github_code", "test-ua", 0)
}

func TestGitHubCode_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "golang", r.URL.Query().Get("q"))
		assert.Equal(t, "application/vnd.github.text-match+json", r.Header.Get("Accept"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"name": "main.go",
					"path": "cmd/main.go",
					"html_url": "https://github.com/org/repo/blob/main/cmd/main.go",
					"text_matches": []map[string]interface{}{
						{
							"object_type": "FileContent",
							"property":    "content",
							"fragment":    "package main\n\nfunc main() {}",
							"matches": []map[string]interface{}{
								{"indices": []int{8, 12}},
							},
						},
					},
					"repository": map[string]interface{}{
						"full_name":   "org/repo",
						"html_url":    "https://github.com/org/repo",
						"description": "repo desc",
					},
				},
			},
		})
	}))
	defer server.Close()

	old := searchAPI
	searchAPI = server.URL
	defer func() { searchAPI = old }()

	engine.Reset()
	engine.Register("github_code", &GitHubCode{})
	eng, _ := engine.Get("github_code")

	require.True(t, eng.Setup(engine.EngineInitConfig{
		Client: testClient(t),
		Extra: map[string]any{
			"ghc_auth": map[string]any{"type": "none"},
		},
	}))

	resp, err := eng.Search(context.Background(), &models.Request{
		Query:    "golang",
		Category: models.CategoryIT,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "code", resp.Results[0].Kind)
	assert.Equal(t, "code.html", resp.Results[0].Template)
	require.Len(t, resp.TypedResults, 1)
	cr, ok := resp.TypedResults[0].(*results.CodeResult)
	require.True(t, ok)
	assert.Equal(t, "cmd/main.go", cr.Filename)
	assert.Equal(t, "https://github.com/org/repo", cr.Repository)
}
