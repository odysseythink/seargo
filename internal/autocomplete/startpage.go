package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewStartpageProvider(client *httpx.Client) Provider {
	return &startpageProvider{client: client}
}

type startpageProvider struct {
	client *httpx.Client
}

func (p *startpageProvider) Fetch(ctx context.Context, query string, locale string) ([]string, error) {
	lang := LocaleToStartpageLanguage(locale)

	resp, err := p.client.R().
		SetQueryParam("q", query).
		SetQueryParam("language", lang).
		SetQueryParam("format", "json").
		SetContext(ctx).
		Get("https://www.startpage.com/sp/search/suggestions")
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
