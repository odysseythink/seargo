package processor

import (
	"context"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/search/query"
	"github.com/seargo/seargo/pkg/models"
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
	req := &models.Request{
		Query:      params.Query,
		Language:   params.Language,
		SafeSearch: params.SafeSearch,
		TimeRange:  params.TimeRange,
		Page:       params.PageNo,
	}
	resp, err := p.eng.Search(ctx, req)
	if err != nil {
		p.RecordResult(false, err)
		return nil, err
	}
	p.RecordResult(true, nil)
	return &ProcessorResult{
		Results:     resp.Results,
		Suggestions: resp.Suggestions,
	}, nil
}
