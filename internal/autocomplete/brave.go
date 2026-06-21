package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewBraveProvider(client *httpx.Client) Provider {
	return &braveProvider{client: client}
}

type braveProvider struct {
	client *httpx.Client
}

func (p *braveProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("q", query).
		SetQueryParam("rich", "false").
		SetQueryParam("source", "web").
		SetHeader("Cookie", "sup=1").
		SetContext(ctx).
		Get("https://search.brave.com/api/suggest")
	if err != nil {
		return nil, err
	}

	var data struct {
		Results []struct {
			Query string `json:"query"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}

	results := make([]string, 0, len(data.Results))
	for _, r := range data.Results {
		if r.Query != "" {
			results = append(results, r.Query)
		}
	}
	return results, nil
}
