package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewSeznamProvider(client *httpx.Client) Provider {
	return &seznamProvider{client: client}
}

type seznamProvider struct {
	client *httpx.Client
}

func (p *seznamProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("q", query).
		SetContext(ctx).
		Get("https://suggest.fulltext.seznam.cz/fulltext")
	if err != nil {
		return nil, err
	}

	var data []interface{}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}
	if len(data) < 2 {
		return nil, nil
	}

	phrases, ok := data[1].([]interface{})
	if !ok {
		return nil, nil
	}

	results := make([]string, 0, len(phrases))
	for _, p := range phrases {
		item, ok := p.([]interface{})
		if !ok || len(item) == 0 {
			continue
		}
		if text, ok := item[0].(string); ok && text != "" {
			results = append(results, text)
		}
	}
	return results, nil
}
