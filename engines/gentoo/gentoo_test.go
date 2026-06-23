package gentoo

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
	return httpx.NewClient(reg, "", "gentoo", "test-ua", 0)
}

func TestGentoo_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasSuffix(r.URL.Path, "/api.php"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"query": map[string]interface{}{
				"search": []map[string]interface{}{
					{"title": "Ebuild", "pageid": 1, "snippet": "..."},
				},
			},
		})
	}))
	defer server.Close()

	engine.Reset()
	engine.Register("gentoo", &Gentoo{})
	eng, _ := engine.Get("gentoo")

	require.True(t, eng.Setup(engine.EngineInitConfig{
		Client:     testClient(t),
		Categories: []models.Category{models.CategoryIT},
		Extra: map[string]any{
			"base_url": server.URL + "/",
			"api_path": "api.php",
		},
	}))

	resp, err := eng.Search(context.Background(), &models.Request{
		Query:    "ebuild",
		Category: models.CategoryIT,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "Ebuild", resp.Results[0].Title)
	assert.Contains(t, resp.Results[0].URL, "/wiki/Ebuild")
}
