package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewQihoo360Provider(client *httpx.Client) Provider {
	return &qihoo360Provider{client: client}
}

type qihoo360Provider struct {
	client *httpx.Client
}

func (p *qihoo360Provider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("word", query).
		SetContext(ctx).
		Get("https://sug.so.360.cn/suggest")
	if err != nil {
		return nil, err
	}

	var data struct {
		S []string `json:"s"`
	}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}

	results := make([]string, 0, len(data.S))
	for _, s := range data.S {
		if s != "" {
			results = append(results, s)
		}
	}
	return results, nil
}
