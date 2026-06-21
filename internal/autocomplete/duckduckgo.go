package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewDuckDuckGoProvider(client *httpx.Client) Provider {
	return &duckduckgoProvider{client: client}
}

type duckduckgoProvider struct {
	client *httpx.Client
}

func (p *duckduckgoProvider) Fetch(ctx context.Context, query string, locale string) ([]string, error) {
	region := LocaleToDDGRegion(locale)

	resp, err := p.client.R().
		SetQueryParam("type", "list").
		SetQueryParam("q", query).
		SetQueryParam("kl", region).
		SetContext(ctx).
		Get("https://duckduckgo.com/ac/")
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

	suggestions, ok := data[1].([]interface{})
	if !ok {
		return nil, nil
	}

	results := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		if text, ok := s.(string); ok && text != "" {
			results = append(results, text)
		}
	}
	return results, nil
}
