package search

import (
	"context"
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
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/search/processor"
	"github.com/seargo/seargo/internal/search/query"
	"github.com/seargo/seargo/pkg/models"
)

func TestMain(m *testing.M) {
	flag.Set("logtostderr", "true")
	logger.Init("warn", "stdout")
	os.Exit(m.Run())
}

type mockProcessor struct {
	eng           *mockEngineForSched
	result        *processor.ProcessorResult
	err           error
	suspendedFlag bool
}

func (m *mockProcessor) Engine() engine.Engine { return m.eng }
func (m *mockProcessor) Search(ctx context.Context, q *query.ParsedQuery, page int) (*processor.ProcessorResult, error) {
	return m.result, m.err
}
func (m *mockProcessor) Suspended() bool { return m.suspendedFlag }
func (m *mockProcessor) RecordResult(ok bool, err error) {}
func (m *mockProcessor) GetParams(q *query.ParsedQuery, page int) (*processor.RequestParams, bool) {
	return &processor.RequestParams{Query: "test", PageNo: 1}, true
}

type mockEngineForSched struct {
	name       string
	categories []models.Category
}

func (m *mockEngineForSched) Name() string                            { return m.name }
func (m *mockEngineForSched) Categories() []models.Category           { return m.categories }
func (m *mockEngineForSched) Capabilities() engine.Capabilities       { return engine.Capabilities{} }
func (m *mockEngineForSched) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }
func (m *mockEngineForSched) Setup(cfg engine.EngineInitConfig) bool                  { return true }
func (m *mockEngineForSched) About() engine.EngineAbout                               { return engine.EngineAbout{} }
func (m *mockEngineForSched) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return &models.Response{}, nil
}

func TestIsEngineEnabled(t *testing.T) {
	assert.True(t, isEngineEnabled(config.EngineConfig{Enabled: true, Disabled: true}))
	assert.True(t, isEngineEnabled(config.EngineConfig{Enabled: true, Disabled: false}))
	assert.True(t, isEngineEnabled(config.EngineConfig{Enabled: false, Disabled: false}))
	assert.False(t, isEngineEnabled(config.EngineConfig{Enabled: false, Disabled: true}))
}

func TestSelectProcessors_ByCategory(t *testing.T) {
	s := &Scheduler{
		processors: map[string]processor.Processor{
			"google": &mockProcessor{eng: &mockEngineForSched{name: "google"}},
			"bing":   &mockProcessor{eng: &mockEngineForSched{name: "bing"}, suspendedFlag: true},
		},
		categoriesAsTabs: map[string]config.CategoryTabConfig{
			"general": {Engines: []string{"google", "bing"}},
		},
	}

	selected := s.selectProcessors(&query.ParsedQuery{}, models.CategoryGeneral)
	assert.Len(t, selected, 1, "suspended bing should be excluded")
	assert.Equal(t, "google", selected[0].Engine().Name())
}

func TestSelectProcessors_ByBang(t *testing.T) {
	s := &Scheduler{
		processors: map[string]processor.Processor{
			"github":        &mockProcessor{eng: &mockEngineForSched{name: "github"}},
			"stackoverflow": &mockProcessor{eng: &mockEngineForSched{name: "stackoverflow"}},
		},
	}

	parsed := &query.ParsedQuery{EngineRefs: []string{"github"}}
	selected := s.selectProcessors(parsed, models.CategoryGeneral)
	assert.Len(t, selected, 1)
	assert.Equal(t, "github", selected[0].Engine().Name())
}

func TestComputeTimeout(t *testing.T) {
	s := &Scheduler{
		defaultEngineTimeout: 3 * time.Second,
		globalTimeout:        30 * time.Second,
	}

	procs := []processor.Processor{
		&mockProcessor{eng: &mockEngineForSched{name: "fast"}},
	}
	timeout := s.computeTimeout(&query.ParsedQuery{}, procs)
	assert.Equal(t, 3*time.Second, timeout)
}

func TestCacheKey(t *testing.T) {
	s := &Scheduler{}
	pq := &query.ParsedQuery{
		Terms:      []string{"hello", "world"},
		EngineRefs: []string{"google"},
		Categories: []models.Category{models.CategoryGeneral},
	}
	req := &models.Request{Category: models.CategoryGeneral, SafeSearch: 1, TimeRange: "week", Page: 1, PageSize: 10}

	key1 := s.cacheKey(pq, req)
	key2 := s.cacheKey(pq, req)
	assert.Equal(t, key1, key2, "same params should produce same key")

	req2 := &models.Request{Category: models.CategoryImages, SafeSearch: 1, TimeRange: "week", Page: 1, PageSize: 10}
	key3 := s.cacheKey(pq, req2)
	assert.NotEqual(t, key1, key3, "different category should produce different key")
}

func TestExternalBangURL(t *testing.T) {
	url, ok := externalBangURL("g", []string{"golang"})
	assert.True(t, ok)
	assert.Contains(t, url, "google.com")
	assert.Contains(t, url, "golang")

	_, ok = externalBangURL("nonexistent", []string{"test"})
	assert.False(t, ok)
}

func TestScheduler_ExternalBang(t *testing.T) {
	c, _ := cache.NewMultiLevel("")
	cfg := &config.Config{
		Search:   config.SearchConfig{MaxResults: 10, SafeSearch: 1},
		Engines:  []config.EngineConfig{},
		Outgoing: config.OutgoingConfig{RequestTimeout: 15},
	}

	s, err := NewScheduler(cfg, c, nil)
	require.NoError(t, err)

	resp, err := s.Search(context.Background(), &models.Request{
		Query:    "!!g golang",
		Category: models.CategoryGeneral,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.RedirectURL)
	assert.Contains(t, resp.RedirectURL, "google.com")
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

	window2, _ := paginate(results, 3, 10)
	assert.Len(t, window2, 5)
}
