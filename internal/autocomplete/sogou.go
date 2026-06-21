package autocomplete

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/seargo/seargo/internal/httpx"
)

func NewSogouProvider(client *httpx.Client) Provider {
	return &sogouProvider{client: client}
}

type sogouProvider struct {
	client *httpx.Client
}

func (p *sogouProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("query", query).
		SetContext(ctx).
		Get("https://waptest.sogou.com/api/search/suggest")
	if err != nil {
		return nil, err
	}

	// Extract JSON array from response, handling any callback padding
	body := string(resp.Body)
	start := strings.Index(body, "[")
	end := strings.LastIndex(body, "]")
	if start < 0 || end < 0 || start >= end {
		return nil, nil
	}
	rawJSON := body[start : end+1]

	var data []interface{}
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return nil, err
	}
	if len(data) < 2 {
		return nil, nil
	}

	suggestions, ok := data[1].([]interface{})
	if !ok {
		return nil, nil
	}

	results := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		item, ok := s.([]interface{})
		if !ok || len(item) == 0 {
			continue
		}
		if text, ok := item[0].(string); ok && text != "" {
			results = append(results, text)
		}
	}
	return results, nil
}
