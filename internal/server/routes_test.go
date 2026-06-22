package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/metrics"
	"github.com/seargo/seargo/internal/search"
)

func TestHandleSearchFormatDisallowed406(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := &Server{
		router: r,
		config: &config.Config{
			Search: config.SearchConfig{
				Formats: []string{"json"},
			},
		},
		scheduler:         nil,
		enginesStatsStore: metrics.NewEngineStatsStore(100),
	}
	r.GET("/api/search", s.handleSearch)

	req := httptest.NewRequest("GET", "/api/search?q=test&format=csv", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 406 {
		t.Errorf("expected 406 for disallowed format, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSearchFormatUnknown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := &Server{
		router: r,
		config: &config.Config{
			Search: config.SearchConfig{
				Formats: []string{"html", "json", "rss"},
			},
		},
		scheduler:         nil,
		enginesStatsStore: metrics.NewEngineStatsStore(100),
	}
	r.GET("/api/search", s.handleSearch)

	req := httptest.NewRequest("GET", "/api/search?q=test&format=pdf", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 406 {
		t.Errorf("expected 406 for unknown format, got %d", w.Code)
	}
}

func TestStatsRoutesRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test with EnableMetrics=true — routes should be registered
	s := &Server{
		router: gin.New(),
		config: &config.Config{
			General: config.GeneralConfig{EnableMetrics: true},
		},
		enginesStatsStore: metrics.NewEngineStatsStore(100),
	}
	s.setupRoutes()

	req := httptest.NewRequest("GET", "/api/stats/engines", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200 with EnableMetrics=true, got %d: %s", w.Code, w.Body.String())
	}
}

func TestStatsRoutesGatedWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test with EnableMetrics=false — routes should NOT be registered (404)
	s := &Server{
		router: gin.New(),
		config: &config.Config{
			General: config.GeneralConfig{EnableMetrics: false},
		},
		enginesStatsStore: metrics.NewEngineStatsStore(100),
	}
	s.setupRoutes()

	req := httptest.NewRequest("GET", "/api/stats/engines", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404 with EnableMetrics=false, got %d", w.Code)
	}
}

func TestHandleSearchResponseKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Search: config.SearchConfig{
			DefaultLang:     "en",
			DefaultCategory: "general",
			MaxResults:      10,
			Formats:         []string{"html", "json"},
		},
		CategoriesAsTabs: map[string]config.CategoryTabConfig{
			"general": {},
		},
	}
	scheduler, err := search.NewScheduler(cfg, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create scheduler: %v", err)
	}

	s := &Server{
		router:    gin.New(),
		config:    cfg,
		scheduler: scheduler,
	}
	s.router.GET("/api/search", s.handleSearch)

	req := httptest.NewRequest("GET", "/api/search?q=test&format=json", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	required := []string{
		"query", "results", "suggestions", "answers", "corrections",
		"infoboxes", "engine_data", "total", "page", "page_size",
		"engines_used", "engines_failed", "response_time_ms",
	}
	for _, key := range required {
		if _, ok := resp[key]; !ok {
			t.Errorf("response missing required key %q", key)
		}
	}
}
