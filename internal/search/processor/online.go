package processor

import (
	"context"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/search/query"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

type OnlineProcessor struct {
	BaseProcessor
	eng    engine.Engine
	client *httpx.Client
}

func NewOnlineProcessor(eng engine.Engine, suspension Suspension, client *httpx.Client) *OnlineProcessor {
	return &OnlineProcessor{
		BaseProcessor: BaseProcessor{engineName: eng.Name(), suspension: suspension},
		eng:           eng,
		client:        client,
	}
}

func (p *OnlineProcessor) Engine() engine.Engine { return p.eng }

func (p *OnlineProcessor) GetParams(q *query.ParsedQuery, page int) (*RequestParams, bool) {
	caps := p.eng.Capabilities()
	if page > 1 && !caps.SupportsPagination {
		return nil, false
	}
	if q.TimeRange != "" && !caps.SupportsTimeRange {
		return nil, false
	}
	return &RequestParams{
		Query:      strings.Join(q.Terms, " "),
		PageNo:     page,
		Language:   q.Lang,
		TimeRange:  q.TimeRange,
		SafeSearch: q.SafeSearch,
	}, true
}

func (p *OnlineProcessor) Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error) {
	params, ok := p.GetParams(q, page)
	if !ok {
		return nil, ErrUnsupportedSearch
	}

	// Extract resolved locale from context (set by scheduler in executeProcessors)
	if resolved, ok := ctx.Value(CtxKeyResolvedLocale).(engine.ResolvedLocale); ok {
		params.ResolvedLocale = resolved
	}

	req := &models.Request{
		Query:      params.Query,
		Language:   params.Language,
		SafeSearch: params.SafeSearch,
		TimeRange:  params.TimeRange,
		Page:       params.PageNo,
	}
	// Apply engine-specific resolved locale from traits
	if params.ResolvedLocale.Language != "" {
		req.Language = params.ResolvedLocale.Language
	}
	if params.ResolvedLocale.Region != "" {
		req.Locale = params.ResolvedLocale.Region
	}
	if cat, ok := ctx.Value(CtxKeySearchCategory).(models.Category); ok {
		req.Category = cat
	}
	resp, err := p.eng.Search(ctx, req)
	if err != nil {
		p.RecordResult(false, err)
		return nil, err
	}
	p.RecordResult(true, nil)

	var typedResults []results.Result
	if len(resp.TypedResults) > 0 {
		typedResults = make([]results.Result, 0, len(resp.TypedResults))
		for _, raw := range resp.TypedResults {
			if r, ok := raw.(results.Result); ok {
				typedResults = append(typedResults, r)
			}
		}
	}
	if len(typedResults) == 0 {
		typedResults = make([]results.Result, 0, len(resp.Results))
		for _, r := range resp.Results {
			typedResults = append(typedResults, results.WrapAPIMainResult(r))
		}
	}

	return &ProcessorResult{
		Results:      resp.Results,
		TypedResults: typedResults,
		Suggestions:  resp.Suggestions,
		Answers:      resp.Answers,
		Corrections:  resp.Corrections,
		Infoboxes:    resp.Infoboxes,
		EngineData:   resp.EngineData,
	}, nil
}
