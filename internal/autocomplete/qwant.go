package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewQwantProvider(client *httpx.Client) Provider {
	return &qwantProvider{client: client}
}

type qwantProvider struct {
	client *httpx.Client
}

func (p *qwantProvider) Fetch(ctx context.Context, query string, locale string) ([]string, error) {
	qwantLocale := LocaleToQwantLocale(locale)

	resp, err := p.client.R().
		SetQueryParam("q", query).
		SetQueryParam("locale", qwantLocale).
		SetContext(ctx).
		Get("https://api.qwant.com/v3/suggest")
	if err != nil {
		return nil, err
	}

	var data struct {
		Status string `json:"status"`
		Data   struct {
			Items []struct {
				Value string `json:"value"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}
	if data.Status != "success" {
		return nil, nil
	}

	results := make([]string, 0, len(data.Data.Items))
	for _, item := range data.Data.Items {
		if item.Value != "" {
			results = append(results, item.Value)
		}
	}
	return results, nil
}
