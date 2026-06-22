package server

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/cache"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/storage"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/search"
	"github.com/seargo/seargo/pkg/models"
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
	c := makeTestCache(t)
	sched, _ := search.NewScheduler(cfg, c, nil, nil, nil, nil, nil)

	srv := New(cfg, sched, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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
	c := makeTestCache(t)
	sched, _ := search.NewScheduler(cfg, c, nil, nil, nil, nil, nil)

	srv := New(cfg, sched, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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
		Server: config.ServerConfig{
			Port:            8080,
			BindAddress:     "0.0.0.0",
			PublicInstance:  false,
			Method:          "POST",
			SecretKey:       "super-secret-do-not-leak",
		},
		Search: config.SearchConfig{
			DefaultLang:     "zh-CN",
			DefaultCategory: "general",
			SafeSearch:      1,
			Autocomplete:    "google",
			MaxResults:      10,
		},
		General: config.GeneralConfig{
			InstanceName:  "TestInstance",
			Debug:         false,
			EnableMetrics: true,
		},
		UI: config.UIConfig{
			DefaultTheme:  "simple",
			DefaultLocale: "",
			Hotkeys:       "default",
		},
		Outgoing: config.OutgoingConfig{RequestTimeout: 15},
		Engines: []config.EngineConfig{
			{Name: "google", Engine: "google", APIKey: "secret-api-key"},
		},
	}
	c := makeTestCache(t)
	sched, _ := search.NewScheduler(cfg, c, nil, nil, nil, nil, nil)

	srv := New(cfg, sched, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config", nil)
	srv.router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	body := w.Body.String()

	// Present: public config fields
	assert.Contains(t, body, "zh-CN")
	assert.Contains(t, body, "TestInstance")
	assert.Contains(t, body, "google")

	// Absent: secrets MUST NOT leak
	assert.NotContains(t, body, "super-secret-do-not-leak")
	assert.NotContains(t, body, "secret-api-key")
	assert.NotContains(t, body, "SecretKey")
	assert.NotContains(t, body, "APIKey")
	assert.NotContains(t, body, "secret_key")
}

type mockEngineForServer struct {
	name       string
	categories []models.Category
}

func (m *mockEngineForServer) Name() string                       { return m.name }
func (m *mockEngineForServer) Categories() []models.Category      { return m.categories }
func (m *mockEngineForServer) Capabilities() engine.Capabilities  { return engine.Capabilities{} }
func (m *mockEngineForServer) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }
func (m *mockEngineForServer) Setup(cfg engine.EngineInitConfig) bool                    { return true }
func (m *mockEngineForServer) About() engine.EngineAbout                                  { return engine.EngineAbout{} }
func (m *mockEngineForServer) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return &models.Response{}, nil
}

func TestEnginesEndpoint(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, BindAddress: "0.0.0.0"},
		Search: config.SearchConfig{MaxResults: 10},
		Engines: []config.EngineConfig{
			{Name: "Google", Engine: "google", Shortcut: "g", Disabled: false, Weight: 1.0},
			{Name: "Bing", Engine: "bing", Shortcut: "b", Disabled: true, Weight: 1.0},
		},
		Outgoing: config.OutgoingConfig{RequestTimeout: 15},
	}
	mockEngine := &mockEngineForServer{name: "google", categories: []models.Category{models.CategoryGeneral}}
	engine.Register("google", mockEngine)
	c := makeTestCache(t)
	sched, _ := search.NewScheduler(cfg, c, nil, nil, nil, nil, nil)

	srv := New(cfg, sched, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/engines", nil)
	srv.router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "google")
	assert.Contains(t, body, `"enabled":true`)
	assert.NotContains(t, body, `"name":"bing"`)
}

func makeTestCache(t *testing.T) cache.Cache {
	t.Helper()
	kv, err := storage.New(storage.Options{Backend: "memory", NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { kv.Close() })
	c, err := cache.NewMultiLevel(kv, cache.Config{Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

