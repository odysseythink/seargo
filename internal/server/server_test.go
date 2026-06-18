package server

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/cache"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/search"
)

func TestMain(m *testing.M) {
	flag.Set("logtostderr", "true")
	logger.Init("info", "stdout")
	os.Exit(m.Run())
}

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, BindAddress: "0.0.0.0"},
		Search: config.SearchConfig{DefaultLang: "zh-CN"},
	}
	c, _ := cache.NewMultiLevel("")
	sched, _ := search.NewScheduler(cfg, c)

	srv := New(cfg, sched)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	srv.router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestCategoriesEndpoint(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, BindAddress: "0.0.0.0"},
		Search: config.SearchConfig{MaxResults: 10},
		CategoriesAsTabs: map[string]config.CategoryTabConfig{
			"general": {Engines: []string{"google", "bing"}},
			"images":  {Engines: []string{"google"}},
			"news":    {},
		},
		Engines: []config.EngineConfig{
			{Name: "google", Engine: "google", Categories: []string{"general", "images"}, Weight: 1.0},
			{Name: "bing", Engine: "bing", Categories: []string{"general"}, Weight: 1.0},
		},
		Outgoing: config.OutgoingConfig{RequestTimeout: 15},
	}
	c, _ := cache.NewMultiLevel("")
	sched, _ := search.NewScheduler(cfg, c)

	srv := New(cfg, sched)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/categories", nil)
	srv.router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "general")
	assert.Contains(t, body, "images")
	assert.Contains(t, body, "news")
	assert.NotContains(t, body, `"videos"`)
}

func TestConfigEndpoint(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Search: config.SearchConfig{DefaultLang: "zh-CN", DefaultCategory: "general"},
	}
	c, _ := cache.NewMultiLevel("")
	sched, _ := search.NewScheduler(cfg, c)

	srv := New(cfg, sched)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config", nil)
	srv.router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "zh-CN")
}
