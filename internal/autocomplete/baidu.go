package autocomplete

import (
	"context"
	"encoding/json"

	"github.com/seargo/seargo/internal/httpx"
)

func NewBaiduProvider(client *httpx.Client) Provider {
	return &baiduProvider{client: client}
}

type baiduProvider struct {
	client *httpx.Client
}

func (p *baiduProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("wd", query).
		SetQueryParam("tn", "baidu").
		SetContext(ctx).
		Get("https://suggestion.baidu.com/su")
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
