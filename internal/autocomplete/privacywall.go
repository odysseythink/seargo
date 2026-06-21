package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewPrivacyWallProvider(client *httpx.Client) Provider {
	return &privacywallProvider{client: client}
}

type privacywallProvider struct {
	client *httpx.Client
}

func (p *privacywallProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("q", query).
		SetQueryParam("format", "json").
		SetContext(ctx).
		Get("https://api.privacywall.org/suggest")
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
