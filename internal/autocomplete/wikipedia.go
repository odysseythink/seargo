package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewWikipediaProvider(client *httpx.Client) Provider {
	return &wikipediaProvider{client: client}
}

type wikipediaProvider struct {
	client *httpx.Client
}

func (p *wikipediaProvider) Fetch(ctx context.Context, query string, locale string) ([]string, error) {
	netloc := LocaleToWikipediaNetloc(locale)

	resp, err := p.client.R().
		SetQueryParam("action", "opensearch").
		SetQueryParam("search", query).
		SetQueryParam("limit", "10").
		SetQueryParam("namespace", "0").
		SetQueryParam("format", "json").
		SetContext(ctx).
		Get("https://" + netloc + "/w/api.php")
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
