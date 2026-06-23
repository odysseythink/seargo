package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/seargo/seargo/pkg/models"
)

// SearchParams are the common query knobs used by parity tests.
type SearchParams struct {
	Category   string
	Language   string
	Locale     string
	SafeSearch int
	TimeRange  string
	Page       int
	PageSize   int
}

func (p SearchParams) toUpstream() url.Values {
	v := url.Values{}
	if p.Category != "" {
		// SearXNG selects categories via category_<name>=1, not ?category=...
		v.Set("category_"+p.Category, "1")
	}
	if p.Language != "" {
		v.Set("language", p.Language)
	}
	if p.SafeSearch > 0 {
		v.Set("safesearch", fmt.Sprintf("%d", p.SafeSearch))
	}
	if p.TimeRange != "" {
		v.Set("time_range", p.TimeRange)
	}
	if p.Page > 0 {
		v.Set("pageno", fmt.Sprintf("%d", p.Page))
	}
	return v
}

func (p SearchParams) toSearGo() url.Values {
	v := url.Values{}
	if p.Category != "" {
		v.Set("category", p.Category)
	}
	if p.Language != "" {
		v.Set("language", p.Language)
	}
	if p.Locale != "" {
		v.Set("locale", p.Locale)
	}
	if p.SafeSearch > 0 {
		v.Set("safesearch", fmt.Sprintf("%d", p.SafeSearch))
	}
	if p.TimeRange != "" {
		v.Set("time_range", p.TimeRange)
	}
	if p.Page > 0 {
		v.Set("page", fmt.Sprintf("%d", p.Page))
	}
	if p.PageSize > 0 {
		v.Set("page_size", fmt.Sprintf("%d", p.PageSize))
	}
	return v
}

// Client talks to SearGo and upstream SearXNG.
type Client struct {
	http   *http.Client
	config *Config
}

// NewClient creates a parity-test client.
func NewClient(cfg *Config) *Client {
	return &Client{
		http:   &http.Client{Timeout: 30 * time.Second},
		config: cfg,
	}
}

// SearchUpstream calls the upstream SearXNG `/search?format=json` endpoint.
func (c *Client) SearchUpstream(ctx context.Context, query string, params SearchParams) (*UpstreamResponse, error) {
	values := params.toUpstream()
	values.Set("q", query)
	values.Set("format", "json")
	u := c.config.UpstreamBaseURL + "/search?" + values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	return c.doUpstream(req)
}

// SearchSearGo calls the SearGo `/api/search` endpoint.
func (c *Client) SearchSearGo(ctx context.Context, query string, params SearchParams) (*models.Response, error) {
	values := params.toSearGo()
	values.Set("q", query)
	u := c.config.SearGoBaseURL + "/api/search?" + values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	return c.doSearGo(req)
}

func (c *Client) doUpstream(req *http.Request) (*UpstreamResponse, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var up UpstreamResponse
	if err := json.Unmarshal(body, &up); err != nil {
		return nil, fmt.Errorf("decode upstream: %w", err)
	}
	return &up, nil
}

func (c *Client) doSearGo(req *http.Request) (*models.Response, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("seargo returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var sg models.Response
	if err := json.Unmarshal(body, &sg); err != nil {
		return nil, fmt.Errorf("decode seargo: %w", err)
	}
	return &sg, nil
}

// WaitForReady polls baseURL/path until it returns 200 or ctx is cancelled.
func (c *Client) WaitForReady(ctx context.Context, baseURL, path string) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	url := baseURL + path
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := c.http.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// UpstreamExternalBangURL returns the Location header for an `!!bang` query.
func (c *Client) UpstreamExternalBangURL(ctx context.Context, bang, query string) (string, error) {
	u := c.config.UpstreamBaseURL + "/search?q=" + url.QueryEscape("!!"+bang+" "+query)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	loc := resp.Header.Get("Location")
	if loc == "" {
		return "", fmt.Errorf("upstream did not return a redirect")
	}
	return loc, nil
}
