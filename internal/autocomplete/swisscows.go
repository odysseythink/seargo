package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewSwisscowsProvider(client *httpx.Client) Provider {
	return &swisscowsProvider{client: client}
}

type swisscowsProvider struct {
	client *httpx.Client
}

func (p *swisscowsProvider) Fetch(ctx context.Context, query string, locale string) ([]string, error) {
	region := LocaleToDDGRegion(locale)

	resp, err := p.client.R().
		SetQueryParam("query", query).
		SetQueryParam("region", region).
		SetContext(ctx).
		Get("https://swisscows.com/suggest")
	if err != nil {
		return nil, err
	}

	var data []string
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}

	results := make([]string, 0, len(data))
	for _, s := range data {
		if s != "" {
			results = append(results, s)
		}
	}
	return results, nil
}
