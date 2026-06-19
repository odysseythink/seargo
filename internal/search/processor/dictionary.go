package processor

import (
	"context"
	"regexp"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/search/query"
)

// dictionaryRegex matches queries like "define golang" or "definition of golang"
var dictionaryRegex = regexp.MustCompile(`(?i)^(?:define|definition\s+of)\s+(.+)$`)

// OnlineDictionaryProcessor 处理词典查询。
type OnlineDictionaryProcessor struct {
	BaseProcessor
	client *httpx.Client
}

func NewOnlineDictionaryProcessor(suspension Suspension, client *httpx.Client) *OnlineDictionaryProcessor {
	return &OnlineDictionaryProcessor{
		BaseProcessor: BaseProcessor{engineName: "dictionary", suspension: suspension},
		client:        client,
	}
}

func (p *OnlineDictionaryProcessor) Engine() engine.Engine { return nil }

func (p *OnlineDictionaryProcessor) GetParams(q *query.ParsedQuery, page int) (*RequestParams, bool) {
	query := strings.Join(q.Terms, " ")
	matches := dictionaryRegex.FindStringSubmatch(query)
	if matches == nil {
		return nil, false
	}
	return &RequestParams{Query: matches[1], PageNo: page}, true
}

func (p *OnlineDictionaryProcessor) Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error) {
	_, ok := p.GetParams(q, page)
	if !ok {
		return nil, ErrUnsupportedSearch
	}
	return &ProcessorResult{}, nil
}
