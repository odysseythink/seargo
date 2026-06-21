package autocomplete

import (
	"context"
	"encoding/xml"
	"strings"

	"github.com/seargo/seargo/internal/httpx"
)

func NewDBpediaProvider(client *httpx.Client) Provider {
	return &dbpediaProvider{client: client}
}

type dbpediaProvider struct {
	client *httpx.Client
}

func (p *dbpediaProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	resp, err := p.client.R().
		SetQueryParam("Query", query).
		SetQueryParam("MaxHits", "10").
		SetContext(ctx).
		Get("https://lookup.dbpedia.org/api/search/KeywordSearch")
	if err != nil {
		return nil, err
	}

	type Result struct {
		Label string `xml:"Label"`
	}

	type ArrayOfResults struct {
		XMLName xml.Name `xml:"ArrayOfResults"`
		Results []Result `xml:"Result"`
	}

	var arr ArrayOfResults
	if err := xml.Unmarshal(resp.Body, &arr); err != nil {
		return nil, err
	}

	results := make([]string, 0, len(arr.Results))
	for _, r := range arr.Results {
		label := strings.TrimSpace(r.Label)
		if label != "" {
			results = append(results, label)
		}
	}
	return results, nil
}
