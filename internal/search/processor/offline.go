package processor

import (
	"context"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/search/query"
)

// ValueError 表示一个不会导致暂停的"正常"处理器错误（如没有结果）。
type ValueError struct {
	Message  string
	Original error
}

func (e *ValueError) Error() string {
	return e.Message
}

// OfflineProcessor 处理离线搜索引擎（本地文件系统、DB 等）。
// ValueError 类型的错误会被静默处理，不会导致暂停。
type OfflineProcessor struct {
	BaseProcessor
	eng engine.Engine
}

func NewOfflineProcessor(eng engine.Engine, suspension Suspension) *OfflineProcessor {
	return &OfflineProcessor{
		BaseProcessor: BaseProcessor{engineName: eng.Name(), suspension: suspension},
		eng:           eng,
	}
}

func (p *OfflineProcessor) Engine() engine.Engine { return p.eng }

func (p *OfflineProcessor) GetParams(q *query.ParsedQuery, page int) (*RequestParams, bool) {
	if len(q.Terms) == 0 {
		return nil, false
	}
	return &RequestParams{Query: q.Terms[0], PageNo: page}, true
}

func (p *OfflineProcessor) Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error) {
	params, ok := p.GetParams(q, page)
	if !ok {
		return nil, ErrUnsupportedSearch
	}
	_ = params
	// Offline engines would execute here
	return &ProcessorResult{}, nil
}
