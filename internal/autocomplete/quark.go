package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewQuarkProvider(client *httpx.Client) Provider {
	return &quarkProvider{client: client}
}

type quarkProvider struct {
	client *httpx.Client
}

func (p *quarkProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("q", query).
		SetContext(ctx).
		Get("https://suggest.quark.cn/suggest")
	if err != nil {
		return nil, err
	}

	var data struct {
		Data []struct {
			Topic string `json:"topic"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}

	results := make([]string, 0, len(data.Data))
	for _, d := range data.Data {
		if d.Topic != "" {
			results = append(results, d.Topic)
		}
	}
	return results, nil
}
