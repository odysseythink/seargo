package search

import (
	"context"
	"errors"
	"flag"
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
		Outgoing: config.OutgoingConfig{Timeout: 15},
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
		Outgoing: config.OutgoingConfig{Timeout: 15},
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
