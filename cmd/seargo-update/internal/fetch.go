package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the HTTP client interface used by all update commands.
// The standard *http.Client satisfies this interface, and tests can
// inject an httptest.Server via a custom Client or by overriding URLs.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Helper provides simple HTTP helpers on top of Client.
type Helper struct {
	client Client
}

// New returns a Helper backed by the provided Client, or http.DefaultClient
// when nil.
func New(client Client) *Helper {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Helper{client: client}
}

// Get performs an HTTP GET and returns the response body. It returns an error
// for non-200 status codes.
func (h *Helper) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "SearGo-update/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// PostForm performs an HTTP POST with form-encoded data and returns the body.
func (h *Helper) PostForm(ctx context.Context, url string, data url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/sparql-results+json")
	req.Header.Set("User-Agent", "SearGo-update/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}
