package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewYandexProvider(client *httpx.Client) Provider {
	return &yandexProvider{client: client}
}

type yandexProvider struct {
	client *httpx.Client
}

func (p *yandexProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("part", query).
		SetQueryParam("client", "search").
		SetContext(ctx).
		Get("https://suggest.yandex.ru/suggest-ya.cgi")
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
