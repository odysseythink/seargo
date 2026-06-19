package processor

import (
	"context"
	"regexp"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/search/query"
)

// currencyRegex matches queries like "1 usd to eur", "100 eur in gbp"
var currencyRegex = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(usd|eur|gbp|cny|jpy|chf|cad|aud|hkd|sgd|krw|inr|mxn|brl)\s+(?:to|in)\s+(usd|eur|gbp|cny|jpy|chf|cad|aud|hkd|sgd|krw|inr|mxn|brl)$`)

// currencySymbolMap maps common currency codes to their symbols.
var currencySymbolMap = map[string]string{
	"usd": "$", "eur": "€", "gbp": "£", "cny": "¥", "jpy": "¥",
	"chf": "Fr", "cad": "CA$", "aud": "A$", "hkd": "HK$", "sgd": "S$",
	"krw": "₩", "inr": "₹", "mxn": "MX$", "brl": "R$",
}

// OnlineCurrencyProcessor 处理货币兑换查询。
type OnlineCurrencyProcessor struct {
	BaseProcessor
	client *httpx.Client
}

func NewOnlineCurrencyProcessor(suspension Suspension, client *httpx.Client) *OnlineCurrencyProcessor {
	return &OnlineCurrencyProcessor{
		BaseProcessor: BaseProcessor{engineName: "currency", suspension: suspension},
		client:        client,
	}
}

func (p *OnlineCurrencyProcessor) Engine() engine.Engine { return nil }

func (p *OnlineCurrencyProcessor) GetParams(q *query.ParsedQuery, page int) (*RequestParams, bool) {
	query := strings.Join(q.Terms, " ")
	if !currencyRegex.MatchString(query) {
		return nil, false
	}
	return &RequestParams{Query: query, PageNo: page}, true
}

func (p *OnlineCurrencyProcessor) Search(ctx context.Context, q *query.ParsedQuery, page int) (*ProcessorResult, error) {
	_, ok := p.GetParams(q, page)
	if !ok {
		return nil, ErrUnsupportedSearch
	}
	// TODO: Implement actual currency conversion in Phase 3
	return &ProcessorResult{}, nil
}
