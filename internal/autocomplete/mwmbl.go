package autocomplete

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/seargo/seargo/internal/httpx"
)

func NewMwmblProvider(client *httpx.Client) Provider {
	return &mwmblProvider{client: client}
}

type mwmblProvider struct {
	client *httpx.Client
}

func (p *mwmblProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("q", query).
		SetContext(ctx).
		Get("https://api.mwmbl.org/api/v1/search/suggest")
	if err != nil {
		return nil, err
	}

	var data struct {
		Suggestions []string `json:"suggestions"`
	}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	results := make([]string, 0, len(data.Suggestions))
	for _, s := range data.Suggestions {
		if strings.HasPrefix(strings.ToLower(s), queryLower) && s != "" {
			results = append(results, s)
		}
	}
	return results, nil
}
