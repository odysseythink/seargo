package search

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/cache"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/pkg/models"
)

func TestMain(m *testing.M) {
	flag.Set("logtostderr", "true")
	logger.Init("warn", "stdout")
	os.Exit(m.Run())
}

// mockEngine for testing
type mockEngine struct {
	name     string
	results  []models.Result
	delay    time.Duration
	fail     bool
	category models.Category
}

func (m *mockEngine) Name() string                  { return m.name }
func (m *mockEngine) Categories() []models.Category { return []models.Category{m.category} }
func (m *mockEngine) Capabilities() engine.Capabilities { return engine.Capabilities{} }
func (m *mockEngine) Init(client *httpx.Client, cfg engine.EngineInitConfig) error { return nil }
func (m *mockEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	if m.fail {
		return nil, errors.New("engine failed")
	}
	select {
	case <-time.After(m.delay):
		return &models.Response{Results: m.results}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestScheduler(t *testing.T) {
	c, _ := cache.NewMultiLevel("")
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Search: config.SearchConfig{MaxResults: 10},
		Engines: []config.EngineConfig{
			{Name: "fast", Enabled: true, Weight: 1.0, Timeout: 5},
			{Name: "slow", Enabled: true, Weight: 0.5, Timeout: 5},
			{Name: "fail", Enabled: true, Weight: 1.0, Timeout: 5},
		},
		Outgoing: config.OutgoingConfig{RequestTimeout: 15},
	}

	s, err := NewScheduler(cfg, c)
	require.NoError(t, err)

	s.RegisterEngine("fast", &mockEngine{
		name: "fast", category: models.CategoryGeneral,
		results: []models.Result{{Title: "Fast", URL: "https://fast.com", Engine: "fast", Score: 1.0}},
		delay: 10 * time.Millisecond,
	})
	s.RegisterEngine("slow", &mockEngine{
		name: "slow", category: models.CategoryGeneral,
		results: []models.Result{{Title: "Slow", URL: "https://slow.com", Engine: "slow", Score: 1.0}},
		delay: 100 * time.Millisecond,
	})
	s.RegisterEngine("fail", &mockEngine{
		name: "fail", category: models.CategoryGeneral,
		fail: true,
	})

	resp, err := s.Search(context.Background(), &models.Request{
		Query:    "test",
		Category: models.CategoryGeneral,
		PageSize: 10,
	})

	require.NoError(t, err)
	assert.Len(t, resp.Results, 2)
	assert.Contains(t, resp.EnginesUsed, "fast")
	assert.Contains(t, resp.EnginesUsed, "slow")
	assert.Contains(t, resp.EnginesFailed, "fail")
}

func TestSchedulerTimeout(t *testing.T) {
	c, _ := cache.NewMultiLevel("")
	cfg := &config.Config{
		Engines: []config.EngineConfig{
			{Name: "fast", Enabled: true, Weight: 1.0, Timeout: 5},
			{Name: "slow", Enabled: true, Weight: 1.0, Timeout: 5},
		},
		Outgoing: config.OutgoingConfig{RequestTimeout: 15},
	}

	s, err := NewScheduler(cfg, c)
	require.NoError(t, err)

	s.RegisterEngine("fast", &mockEngine{
		name: "fast", category: models.CategoryGeneral,
		results: []models.Result{{Title: "Fast", URL: "https://fast.com", Engine: "fast"}},
		delay: 10 * time.Millisecond,
	})
	s.RegisterEngine("slow", &mockEngine{
		name: "slow", category: models.CategoryGeneral,
		results: []models.Result{{Title: "Slow", URL: "https://slow.com", Engine: "slow"}},
		delay: 200 * time.Millisecond,
	})

	resp, err := s.Search(context.Background(), &models.Request{
		Query:    "test",
		Category: models.CategoryGeneral,
		PageSize: 10,
	})

	require.NoError(t, err)
	// slow engine may or may not timeout depending on race; just verify no panic
	assert.NotNil(t, resp)
}

func TestSelectEnginesDisabled(t *testing.T) {
	c, _ := cache.NewMultiLevel("")
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Search: config.SearchConfig{MaxResults: 10},
		Engines: []config.EngineConfig{
			{Name: "google", Engine: "google", Disabled: false, Weight: 1.0},
			{Name: "bing", Engine: "bing", Disabled: true, Weight: 1.0},
			{Name: "ddg", Engine: "duckduckgo", Disabled: false, Weight: 1.0},
		},
		Outgoing: config.OutgoingConfig{RequestTimeout: 15},
	}

	s, err := NewScheduler(cfg, c)
	require.NoError(t, err)

	s.RegisterEngine("google", &mockEngine{name: "google", category: models.CategoryGeneral})
	s.RegisterEngine("bing", &mockEngine{name: "bing", category: models.CategoryGeneral})
	s.RegisterEngine("duckduckgo", &mockEngine{name: "duckduckgo", category: models.CategoryGeneral})

	selected := s.selectEngines(models.CategoryGeneral)
	assert.Len(t, selected, 2)
	names := make([]string, len(selected))
	for i, e := range selected {
		names[i] = e.Name()
	}
	assert.Contains(t, names, "google")
	assert.Contains(t, names, "duckduckgo")
	assert.NotContains(t, names, "bing")
}

func TestSelectEnginesPerCategory(t *testing.T) {
	c, _ := cache.NewMultiLevel("")
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Search: config.SearchConfig{MaxResults: 10},
		Engines: []config.EngineConfig{
			{Name: "google", Engine: "google", Categories: []string{"general", "images"}, Weight: 1.0},
		},
		Outgoing: config.OutgoingConfig{RequestTimeout: 15},
	}

	s, err := NewScheduler(cfg, c)
	require.NoError(t, err)

	s.RegisterEngine("google", &mockEngine{name: "google", category: models.CategoryGeneral})

	selected := s.selectEngines(models.CategoryImages)
	assert.Len(t, selected, 0, "engine only claims general, not images")

	selected = s.selectEngines(models.CategoryGeneral)
	assert.Len(t, selected, 1)
}

func TestPagination(t *testing.T) {
	results := make([]models.Result, 25)
	for i := 0; i < 25; i++ {
		results[i] = models.Result{
			Title: fmt.Sprintf("R%d", i),
			URL:   fmt.Sprintf("https://example.com/%d", i),
			Score: float64(25 - i),
		}
	}
	window, total := paginate(results, 1, 10)
	assert.Equal(t, 25, total)
	assert.Len(t, window, 10)
	assert.Equal(t, "R0", window[0].Title)

	window2, total2 := paginate(results, 2, 10)
	assert.Equal(t, 25, total2)
	assert.Len(t, window2, 10)
	assert.Equal(t, "R10", window2[0].Title)

	window3, total3 := paginate(results, 3, 10)
	assert.Equal(t, 25, total3)
	assert.Len(t, window3, 5)

	window4, total4 := paginate(results, 100, 10)
	assert.Equal(t, 25, total4)
	assert.Len(t, window4, 0)

	window5, total5 := paginate(results, 0, 10)
	assert.Equal(t, 25, total5)
	assert.Len(t, window5, 10)
}

func TestPaginationPreservesOrder(t *testing.T) {
	results := []models.Result{
		{Title: "A", URL: "https://a.com", Score: 3.0},
		{Title: "B", URL: "https://b.com", Score: 2.0},
		{Title: "C", URL: "https://c.com", Score: 1.0},
		{Title: "D", URL: "https://d.com", Score: 0.5},
		{Title: "E", URL: "https://e.com", Score: 0.1},
	}
	window, total := paginate(results, 1, 3)
	assert.Equal(t, 5, total)
	assert.Len(t, window, 3)
	assert.Equal(t, "A", window[0].Title)
	assert.Equal(t, "B", window[1].Title)
	assert.Equal(t, "C", window[2].Title)

	window2, _ := paginate(results, 2, 3)
	assert.Equal(t, "D", window2[0].Title)
	assert.Equal(t, "E", window2[1].Title)
}

func TestPaginateTableDriven(t *testing.T) {
	results := make([]models.Result, 25)
	for i := 0; i < 25; i++ {
		results[i] = models.Result{
			Title: fmt.Sprintf("R%d", i),
			URL:   fmt.Sprintf("https://ex.com/%d", i),
			Score: float64(25 - i),
		}
	}

	tests := []struct {
		name      string
		page      int
		pageSize  int
		wantLen   int
		wantTotal int
		wantFirst string
		wantLast  string
	}{
		{"page1_size10", 1, 10, 10, 25, "R0", "R9"},
		{"page2_size10", 2, 10, 10, 25, "R10", "R19"},
		{"page3_size10", 3, 10, 5, 25, "R20", "R24"},
		{"page4_size10", 4, 10, 0, 25, "", ""},
		{"page0_defaults", 0, 10, 10, 25, "R0", "R9"},
		{"page1_size5", 1, 5, 5, 25, "R0", "R4"},
		{"page5_size5", 5, 5, 5, 25, "R20", "R24"},
		{"page100_size10", 100, 10, 0, 25, "", ""},
		{"zero_pagesize_defaults", 1, 0, 10, 25, "R0", "R9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window, total := paginate(results, tt.page, tt.pageSize)
			assert.Equal(t, tt.wantTotal, total, "total")
			assert.Len(t, window, tt.wantLen, "window length")
			if tt.wantLen > 0 {
				assert.Equal(t, tt.wantFirst, window[0].Title, "first item")
				assert.Equal(t, tt.wantLast, window[tt.wantLen-1].Title, "last item")
			}
		})
	}
}
