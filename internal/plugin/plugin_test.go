package plugin

import (
	"testing"

	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/assert"
)

type mockPlugin struct {
	id             string
	info           PluginInfo
	preSearchRet   bool
	onResultRet    bool
	postSearchOut  []models.Result
}

func (m *mockPlugin) ID() string                             { return m.id }
func (m *mockPlugin) Info() PluginInfo                       { return m.info }
func (m *mockPlugin) Init(ctx *AppContext) bool              { return true }
func (m *mockPlugin) PreSearch(ctx *SearchContext) bool      { return m.preSearchRet }
func (m *mockPlugin) OnResult(ctx *SearchContext, r *models.Result) bool { return m.onResultRet }
func (m *mockPlugin) PostSearch(ctx *SearchContext) []models.Result     { return m.postSearchOut }

func TestPluginStorage_RegisterAndGet(t *testing.T) {
	ResetForTest()
	ps := NewPluginStorage()
	p := &mockPlugin{id: "calculator", info: PluginInfo{Name: "Calc"}}

	err := ps.Register(p)
	assert.NoError(t, err)
	assert.Len(t, ps.All(), 1)

	got, ok := ps.Get("calculator")
	assert.True(t, ok)
	assert.Equal(t, "calculator", got.ID())
}

func TestPluginStorage_Register_DuplicateID(t *testing.T) {
	ResetForTest()
	ps := NewPluginStorage()
	ps.Register(&mockPlugin{id: "test"})
	err := ps.Register(&mockPlugin{id: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestPluginStorage_PreSearch_AbortsOnFirstFalse(t *testing.T) {
	ResetForTest()
	ps := NewPluginStorage()
	ps.Register(&mockPlugin{id: "a", preSearchRet: true})
	ps.Register(&mockPlugin{id: "b", preSearchRet: false})
	ps.Register(&mockPlugin{id: "c", preSearchRet: true})

	ctx := &SearchContext{UserPlugins: []string{"a", "b", "c"}}
	ok := ps.PreSearch(ctx)
	assert.False(t, ok)
}

func TestPluginStorage_OnResult_RemovesOnFalse(t *testing.T) {
	ResetForTest()
	ps := NewPluginStorage()
	ps.Register(&mockPlugin{id: "filter", onResultRet: false})

	ctx := &SearchContext{UserPlugins: []string{"filter"}}
	r := &models.Result{Title: "test", URL: "https://example.com"}
	ok := ps.OnResult(ctx, r)
	assert.False(t, ok)
}

func TestPluginStorage_PostSearch_KeywordMatch(t *testing.T) {
	ResetForTest()
	ps := NewPluginStorage()
	p := &mockPlugin{
		id: "time_zone",
		info: PluginInfo{Keywords: []string{"time", "tz"}},
		postSearchOut: []models.Result{{Kind: "answer", Title: "Berlin", Content: "14:30"}},
	}
	ps.Register(p)

	ctx := &SearchContext{Query: "time Berlin", UserPlugins: []string{"time_zone"}}
	results := ps.PostSearch(ctx)
	assert.Len(t, results, 1)

	ctx2 := &SearchContext{Query: "weather Berlin", UserPlugins: []string{"time_zone"}}
	results2 := ps.PostSearch(ctx2)
	assert.Empty(t, results2)
}

func TestPluginStorage_PostSearch_NoKeywordsAlwaysRuns(t *testing.T) {
	ResetForTest()
	ps := NewPluginStorage()
	p := &mockPlugin{
		id: "global_plugin",
		postSearchOut: []models.Result{{Kind: "answer", Title: "always"}},
	}
	ps.Register(p)

	ctx := &SearchContext{Query: "anything", UserPlugins: []string{"global_plugin"}}
	results := ps.PostSearch(ctx)
	assert.Len(t, results, 1)
}

func TestResult_FilterURLs_ReplaceAndRemove(t *testing.T) {
	r := &models.Result{
		URL:          "https://tracker.example.com/page?utm_source=spam",
		ThumbnailURL: "https://cdn.example.com/thumb.jpg",
	}

	calls := 0
	r.FilterURLs(func(res *models.Result, field string, url string) (string, bool) {
		calls++
		if field == "url" && url != "" {
			return "https://clean.example.com/page", true
		}
		if field == "thumbnail_url" {
			return "", false
		}
		return url, true
	})

	assert.Equal(t, "https://clean.example.com/page", r.URL)
	assert.Equal(t, "", r.ThumbnailURL)
	assert.Equal(t, 2, calls)
}

func TestResult_FilterURLs_NoURLFields(t *testing.T) {
	r := &models.Result{Title: "no urls"}
	calls := 0
	r.FilterURLs(func(res *models.Result, field string, url string) (string, bool) {
		calls++
		return url, true
	})
	assert.Equal(t, 0, calls)
}
