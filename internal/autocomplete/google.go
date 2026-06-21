package autocomplete

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/seargo/seargo/internal/httpx"
)

var htmlTagRE = regexp.MustCompile(`<[^>]*>`)

func NewGoogleProvider(client *httpx.Client) Provider {
	return &googleProvider{client: client}
}

type googleProvider struct {
	client *httpx.Client
}

func (p *googleProvider) Fetch(ctx context.Context, query string, locale string) ([]string, error) {
	subdomain := LocaleToGoogleSubdomain(locale)
	hl := LocaleToGoogleHL(locale)

	resp, err := p.client.R().
		SetQueryParam("client", "gws-wiz").
		SetQueryParam("q", query).
		SetQueryParam("hl", hl).
		SetContext(ctx).
		Get("https://" + subdomain + "/complete/search")
	if err != nil {
		return nil, err
	}

	body := string(resp.Body)
	if idx := strings.IndexByte(body, '['); idx >= 0 {
		body = body[idx:]
	}

	var data []interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	group, ok := data[0].([]interface{})
	if !ok {
		return nil, nil
	}

	var results []string
	for _, item := range group {
		inner, ok := item.([]interface{})
		if !ok || len(inner) == 0 {
			continue
		}
		text, ok := inner[0].(string)
		if !ok {
			continue
		}
		text = htmlTagRE.ReplaceAllString(text, "")
		text = strings.TrimSpace(text)
		if text != "" {
			results = append(results, text)
		}
	}
	return results, nil
}
