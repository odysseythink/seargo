package upstream

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_FetchBoth(t *testing.T) {
	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/search", r.URL.Path)
		require.Equal(t, "json", r.URL.Query().Get("format"))
		require.Equal(t, "golang", r.URL.Query().Get("q"))
		_ = json.NewEncoder(w).Encode(UpstreamResponse{Query: "golang"})
	}))
	defer upstreamSrv.Close()

	seargoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/search", r.URL.Path)
		require.Equal(t, "golang", r.URL.Query().Get("q"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"query": "golang"})
	}))
	defer seargoSrv.Close()

	cfg := &Config{SearGoBaseURL: seargoSrv.URL, UpstreamBaseURL: upstreamSrv.URL}
	c := NewClient(cfg)

	up, err := c.SearchUpstream(t.Context(), "golang", SearchParams{})
	require.NoError(t, err)
	require.Equal(t, "golang", up.Query)

	sg, err := c.SearchSearGo(t.Context(), "golang", SearchParams{})
	require.NoError(t, err)
	require.Equal(t, "golang", sg.Query)
}
