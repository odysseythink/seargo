package processor

import (
	"context"
	"errors"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/search/query"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

var ErrUnsupportedSearch = errors.New("unsupported search")

// CtxKeyResolvedLocale is the context key for passing the engine-specific
// resolved locale from the scheduler down to the processor's Search method.
type ctxKeyResolvedLocale struct{}
var CtxKeyResolvedLocale ctxKeyResolvedLocale

// Suspension 定义暂停/恢复能力接口，由 search.SuspensionTracker 实现。
type Suspension interface {
	Ban(engineName, errorClass string)
	IsSuspended(engineName string) bool
}

// RequestParams 是传给底层 engine.Engine.Search 的参数。
type RequestParams struct {
	Query          string
	Category       models.Category
	PageNo         int
	Language       string
	TimeRange      string
	SafeSearch     int
	EngineData     map[string]any
	ResolvedLocale engine.ResolvedLocale // engine-specific locale resolved from traits
}

// ProcessorResult 是单次 processor 搜索返回的结果流。
type ProcessorResult struct {
	Results      []models.Result   // kept during migration
	TypedResults []results.Result  // new
	Suggestions  []string
	Answers      []models.Answer
	Corrections  []string
	Infoboxes    []models.Infobox
	EngineData   map[string]any
}

// Processor 是搜索处理器的统一接口。
type Processor interface {
	Engine() engine.Engine
	Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error)
	Suspended() bool
	RecordResult(ok bool, err error)
	GetParams(q *query.ParsedQuery, page int) (*RequestParams, bool)
}

// BaseProcessor 提供 Suspended/RecordResult 默认实现。
type BaseProcessor struct {
	engineName string
	suspension Suspension
}

func (bp *BaseProcessor) Suspended() bool {
	if bp.suspension == nil {
		return false
	}
	return bp.suspension.IsSuspended(bp.engineName)
}

func (bp *BaseProcessor) RecordResult(ok bool, err error) {
	if ok || bp.suspension == nil {
		return
	}
	bp.suspension.Ban(bp.engineName, classifyError(err))
}

func classifyError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "403") || strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "access denied") {
		return "SearxEngineAccessDenied"
	}
	if strings.Contains(msg, "captcha") || strings.Contains(msg, "recaptcha") ||
		strings.Contains(msg, "challenge") {
		return "SearxEngineCaptcha"
	}
	if strings.Contains(msg, "429") || strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "rate limit") {
		return "SearxEngineTooManyRequests"
	}
	return "SearxEngineTooManyRequests"
}
