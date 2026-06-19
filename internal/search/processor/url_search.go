package processor

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/search/query"
)

// urlRegex matches queries that look like URLs: scheme://... or domain.tld/...
var urlRegex = regexp.MustCompile(`(?i)^(?:https?://|ftp://)?([a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}(?::\d+)?(?:/.*)?$`)

// OnlineURLSearchProcessor 处理 URL 查询（用户输入了一个 URL）。
type OnlineURLSearchProcessor struct {
	BaseProcessor
	client *httpx.Client
}

func NewOnlineURLSearchProcessor(suspension Suspension, client *httpx.Client) *OnlineURLSearchProcessor {
	return &OnlineURLSearchProcessor{
		BaseProcessor: BaseProcessor{engineName: "url_search", suspension: suspension},
		client:        client,
	}
}

func (p *OnlineURLSearchProcessor) Engine() engine.Engine { return nil }

func (p *OnlineURLSearchProcessor) GetParams(q *query.ParsedQuery, page int) (*RequestParams, bool) {
	query := strings.Join(q.Terms, " ")
	if !urlRegex.MatchString(query) {
		return nil, false
	}
	// Add scheme if missing
	if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") && !strings.HasPrefix(query, "ftp://") {
		query = "http://" + query
	}
	if _, err := url.Parse(query); err != nil {
		return nil, false
	}
	return &RequestParams{Query: query, PageNo: page}, true
}

func (p *OnlineURLSearchProcessor) Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error) {
	_, ok := p.GetParams(q, page)
	if !ok {
		return nil, ErrUnsupportedSearch
	}
	return &ProcessorResult{}, nil
}
