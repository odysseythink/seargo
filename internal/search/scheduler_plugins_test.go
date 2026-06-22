package search

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
)

// stubPlugin satisfies plugin.Plugin for scheduler tests.
type stubPlugin struct {
	id                string
	info              plugin.PluginInfo
	preSearchRet      bool
	preSearchCalled   bool
	onResultRet       bool
	onResultCalled    bool
	postSearchResults []models.Result
	postSearchCalled  bool
}

func (s *stubPlugin) ID() string                                     { return s.id }
func (s *stubPlugin) Info() plugin.PluginInfo                        { return s.info }
func (s *stubPlugin) Init(ctx *plugin.AppContext) bool               { return true }
func (s *stubPlugin) PreSearch(ctx *plugin.SearchContext) bool       { s.preSearchCalled = true; return s.preSearchRet }
func (s *stubPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool {
	s.onResultCalled = true
	return s.onResultRet
}
func (s *stubPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	s.postSearchCalled = true
	return s.postSearchResults
}

// stubAnswerer implements answerer.Answerer for testing.
type stubAnswerer struct {
	info    answerer.AnswererInfo
	results []models.Result
}

func (s *stubAnswerer) Keywords() []string                        { return s.info.Keywords }
func (s *stubAnswerer) Info() answerer.AnswererInfo               { return s.info }
func (s *stubAnswerer) Answer(ctx *answerer.AnswerContext) []models.Result { return s.results }

func TestScheduler_PreSearch_AbortsSearch(t *testing.T) {
	plugin.ResetForTest()
	ps := plugin.NewPluginStorage()
	ps.Register(&stubPlugin{id: "abort_plugin", preSearchRet: false})
	as := answerer.NewAnswererStorage()

	cfg := &config.Config{
		Server:  config.ServerConfig{Port: 8080, BindAddress: "127.0.0.1"},
		Search:  config.SearchConfig{DefaultCategory: "general", MaxResults: 10},
		General: config.GeneralConfig{InstanceName: "test"},
	}
	sched, err := NewScheduler(cfg, nil, nil, ps, as, nil, nil)
	require.NoError(t, err)

	resp, err := sched.Search(context.Background(), &models.Request{
		Query: "test query", Page: 1, PageSize: 10,
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Results)
}

func TestScheduler_PreSearch_Passes(t *testing.T) {
	plugin.ResetForTest()
	ps := plugin.NewPluginStorage()
	ps.Register(&stubPlugin{id: "pass_plugin", preSearchRet: true})
	as := answerer.NewAnswererStorage()

	cfg := &config.Config{
		Server:  config.ServerConfig{Port: 8080, BindAddress: "127.0.0.1"},
		Search:  config.SearchConfig{DefaultCategory: "general", MaxResults: 10},
		General: config.GeneralConfig{InstanceName: "test"},
	}
	sched, err := NewScheduler(cfg, nil, nil, ps, as, nil, nil)
	require.NoError(t, err)

	_, err = sched.Search(context.Background(), &models.Request{
		Query: "test query", Page: 1, PageSize: 10,
	})
	if err != nil {
		assert.True(t, strings.Contains(err.Error(), "all engines failed") ||
			strings.Contains(err.Error(), "no processors"),
			"unexpected error: %v", err)
	}
}

func TestScheduler_NilPluginStorage_Noop(t *testing.T) {
	cfg := &config.Config{
		Server:  config.ServerConfig{Port: 8080, BindAddress: "127.0.0.1"},
		Search:  config.SearchConfig{DefaultCategory: "general", MaxResults: 10},
		General: config.GeneralConfig{InstanceName: "test"},
	}
	sched, err := NewScheduler(cfg, nil, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	_, err = sched.Search(context.Background(), &models.Request{
		Query: "test", Page: 1, PageSize: 10,
	})
	if err != nil {
		assert.True(t, strings.Contains(err.Error(), "no processors") ||
			strings.Contains(err.Error(), "all engines failed"),
			"unexpected error: %v", err)
	}
}
