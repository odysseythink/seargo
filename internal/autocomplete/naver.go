package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewNaverProvider(client *httpx.Client) Provider {
	return &naverProvider{client: client}
}

type naverProvider struct {
	client *httpx.Client
}

func (p *naverProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("query", query).
		SetContext(ctx).
		Get("https://ac.search.naver.com/autocomplete")
	if err != nil {
		return nil, err
	}

	var data [][]string
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}
	if len(data) < 2 {
		return nil, nil
	}

	return data[1], nil
}
