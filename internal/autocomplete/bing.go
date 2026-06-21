package autocomplete

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/seargo/seargo/internal/httpx"
)

const (
	bingPUASpan = "\ue000"
	bingPUAEnd  = "\ue001"
)

func NewBingProvider(client *httpx.Client) Provider {
	return &bingProvider{client: client}
}

type bingProvider struct {
	client *httpx.Client
}

func (p *bingProvider) Fetch(ctx context.Context, query string, _ string) ([]string, error) {
	cvid := randomCVID()

	resp, err := p.client.R().
		SetQueryParam("qry", query).
		SetQueryParam("csr", "1").
		SetQueryParam("cvid", cvid).
		SetContext(ctx).
		Get("https://www.bing.com/AS/Suggestions")
	if err != nil {
		return nil, err
	}

	var data struct {
		S []struct {
			Q string `json:"q"`
		} `json:"s"`
	}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, err
	}

	var results []string
	for _, s := range data.S {
		text := s.Q
		text = strings.ReplaceAll(text, bingPUASpan, "")
		text = strings.ReplaceAll(text, bingPUAEnd, "")
		text = strings.TrimSpace(text)
		if text != "" {
			results = append(results, text)
		}
	}
	return results, nil
}

func randomCVID() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}
