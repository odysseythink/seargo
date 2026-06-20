package processor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/search/query"
	"github.com/seargo/seargo/pkg/models"
)

type mockSuspension struct {
	banned map[string]bool
}

func newMockSuspension() *mockSuspension {
	return &mockSuspension{banned: make(map[string]bool)}
}

func (m *mockSuspension) Ban(engineName, errorClass string) {
	m.banned[engineName] = true
}

func (m *mockSuspension) IsSuspended(engineName string) bool {
	return m.banned[engineName]
}

type mockEngine struct {
	name         string
	caps         engine.Capabilities
	searchResult *models.Response
	searchErr    error
}

func (m *mockEngine) Name() string                            { return m.name }
func (m *mockEngine) Categories() []models.Category           { return []models.Category{models.CategoryGeneral} }
func (m *mockEngine) Capabilities() engine.Capabilities       { return m.caps }
func (m *mockEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }
func (m *mockEngine) Setup(cfg engine.EngineInitConfig) bool                    { return true }
func (m *mockEngine) About() engine.EngineAbout                                  { return engine.EngineAbout{} }
func (m *mockEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return m.searchResult, m.searchErr
}

func TestBaseProcessor_RecordResultSuccess(t *testing.T) {
	ms := newMockSuspension()
	bp := &BaseProcessor{engineName: "test", suspension: ms}

	bp.RecordResult(true, nil)
	assert.False(t, ms.IsSuspended("test"), "success should not suspend")
}

func TestBaseProcessor_RecordResultFailure(t *testing.T) {
	ms := newMockSuspension()
	bp := &BaseProcessor{engineName: "test", suspension: ms}

	bp.RecordResult(false, errors.New("403 access denied"))
	assert.True(t, ms.IsSuspended("test"), "failure should suspend")
}

func TestBaseProcessor_Suspended(t *testing.T) {
	ms := newMockSuspension()
	bp := &BaseProcessor{engineName: "test", suspension: ms}

	assert.False(t, bp.Suspended())
	ms.Ban("test", "SearxEngineCaptcha")
	assert.True(t, bp.Suspended())
}

func TestBaseProcessor_RecordResultNilSuspension(t *testing.T) {
	bp := &BaseProcessor{engineName: "test", suspension: nil}
	bp.RecordResult(false, errors.New("err"))
	assert.False(t, bp.Suspended())
}

// --- OnlineProcessor tests ---

func TestOnlineProcessor_GetParams_Default(t *testing.T) {
	eng := &mockEngine{name: "google", caps: engine.Capabilities{SupportsSafeSearch: true, SupportsPagination: true, SupportsTimeRange: true}}
	proc := NewOnlineProcessor(eng, nil, nil)
	q := &query.ParsedQuery{Terms: []string{"hello", "world"}, Lang: "en", SafeSearch: 1, TimeRange: "week"}
	params, ok := proc.GetParams(q, 1)
	assert.True(t, ok)
	assert.Equal(t, "hello world", params.Query)
	assert.Equal(t, 1, params.SafeSearch)
	assert.Equal(t, "week", params.TimeRange)
	assert.Equal(t, "en", params.Language)
}

func TestOnlineProcessor_GetParams_PaginationUnsupported(t *testing.T) {
	eng := &mockEngine{name: "google", caps: engine.Capabilities{SupportsPagination: false}}
	proc := NewOnlineProcessor(eng, nil, nil)
	q := &query.ParsedQuery{Terms: []string{"test"}}
	_, ok := proc.GetParams(q, 2)
	assert.False(t, ok, "page>1 rejected when pagination unsupported")
}

func TestOnlineProcessor_GetParams_TimeRangeUnsupported(t *testing.T) {
	eng := &mockEngine{name: "google", caps: engine.Capabilities{SupportsTimeRange: false}}
	proc := NewOnlineProcessor(eng, nil, nil)
	q := &query.ParsedQuery{Terms: []string{"test"}, TimeRange: "day"}
	_, ok := proc.GetParams(q, 1)
	assert.False(t, ok, "time_range rejected when unsupported")
}

func TestOnlineProcessor_SearchSuccess(t *testing.T) {
	eng := &mockEngine{name: "google", caps: engine.Capabilities{SupportsPagination: true}, searchResult: &models.Response{Results: []models.Result{{Title: "R", URL: "https://x.com"}}, Suggestions: []string{"s1"}}}
	ms := newMockSuspension()
	proc := NewOnlineProcessor(eng, ms, nil)
	q := &query.ParsedQuery{Terms: []string{"test"}}
	res, err := proc.Search(context.Background(), q, 1)
	assert.NoError(t, err)
	assert.Len(t, res.Results, 1)
	assert.Len(t, res.Suggestions, 1)
	assert.False(t, ms.IsSuspended(eng.Name()), "success should not suspend")
}

func TestOnlineProcessor_SearchFailure(t *testing.T) {
	eng := &mockEngine{name: "google", caps: engine.Capabilities{SupportsPagination: true}, searchErr: errors.New("403 forbidden")}
	ms := newMockSuspension()
	proc := NewOnlineProcessor(eng, ms, nil)
	q := &query.ParsedQuery{Terms: []string{"test"}}
	_, err := proc.Search(context.Background(), q, 1)
	assert.Error(t, err)
	assert.True(t, ms.IsSuspended(eng.Name()), "403 should trigger suspension")
}

// --- OfflineProcessor tests ---

func TestOfflineProcessor_GetParams(t *testing.T) {
	eng := &mockEngine{name: "local"}
	proc := NewOfflineProcessor(eng, nil)
	q := &query.ParsedQuery{Terms: []string{"test"}}
	params, ok := proc.GetParams(q, 1)
	assert.True(t, ok)
	assert.Equal(t, "test", params.Query)
}

func TestOfflineProcessor_ValueErrorIgnored(t *testing.T) {
	ms := newMockSuspension()
	eng := &mockEngine{name: "local"}
	proc := NewOfflineProcessor(eng, ms)
	q := &query.ParsedQuery{Terms: []string{"test"}}
	res, err := proc.Search(context.Background(), q, 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, ms.IsSuspended("local"), "ValueError should not suspend")
}

// --- CurrencyProcessor tests ---

func TestCurrencyParser_GetParamsMatch(t *testing.T) {
	proc := NewOnlineCurrencyProcessor(nil, nil)
	tests := []string{
		"1 usd to eur",
		"100 eur in gbp",
		"50.5 cny to usd",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			q := &query.ParsedQuery{Terms: strings.Fields(tt)}
			params, ok := proc.GetParams(q, 1)
			assert.True(t, ok)
			assert.Equal(t, tt, params.Query)
		})
	}
}

func TestCurrencyParser_NoMatch(t *testing.T) {
	proc := NewOnlineCurrencyProcessor(nil, nil)
	tests := []string{
		"golang tutorial",
		"usd to eur", // no amount
		"!!g test",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			q := &query.ParsedQuery{Terms: strings.Fields(tt)}
			_, ok := proc.GetParams(q, 1)
			assert.False(t, ok)
		})
	}
}

// --- DictionaryProcessor tests ---

func TestDictionaryParser_GetParamsMatch(t *testing.T) {
	proc := NewOnlineDictionaryProcessor(nil, nil)
	tests := []struct {
		input string
		word  string
	}{
		{"define golang", "golang"},
		{"definition of algorithm", "algorithm"},
		{"Define Hello", "Hello"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			q := &query.ParsedQuery{Terms: strings.Fields(tt.input)}
			params, ok := proc.GetParams(q, 1)
			assert.True(t, ok)
			assert.Equal(t, tt.word, params.Query)
		})
	}
}

func TestDictionaryParser_NoMatch(t *testing.T) {
	proc := NewOnlineDictionaryProcessor(nil, nil)
	q := &query.ParsedQuery{Terms: []string{"golang", "tutorial"}}
	_, ok := proc.GetParams(q, 1)
	assert.False(t, ok)
}

// --- URLSearchProcessor tests ---

func TestURLSearchParser_GetParamsMatch(t *testing.T) {
	proc := NewOnlineURLSearchProcessor(nil, nil)
	tests := []string{
		"https://example.com",
		"example.com/path",
		"golang.org",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			q := &query.ParsedQuery{Terms: strings.Fields(tt)}
			params, ok := proc.GetParams(q, 1)
			assert.True(t, ok)
			assert.NotEmpty(t, params.Query)
		})
	}
}

func TestURLSearchParser_NoMatch(t *testing.T) {
	proc := NewOnlineURLSearchProcessor(nil, nil)
	q := &query.ParsedQuery{Terms: []string{"golang", "tutorial"}}
	_, ok := proc.GetParams(q, 1)
	assert.False(t, ok)
}

// --- Factory tests ---

func TestNewProcessorFromConfig_Online(t *testing.T) {
	eng := &mockEngine{name: "google"}
	ec := config.EngineConfig{Name: "google", Engine: "google"}
	proc, err := NewProcessorFromConfig(eng, ec, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, proc)
	assert.Equal(t, "google", proc.Engine().Name())
}

func TestNewProcessorFromConfig_NilEngine(t *testing.T) {
	ec := config.EngineConfig{Name: "missing"}
	_, err := NewProcessorFromConfig(nil, ec, nil, nil)
	assert.Error(t, err)
}
