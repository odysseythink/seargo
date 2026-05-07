package httpx

import (
	"time"

	"github.com/imroc/req/v3"
)

type Client struct {
	client *req.Client
}

func New(userAgent string, timeout time.Duration) *Client {
	c := req.C().
		SetUserAgent(userAgent).
		SetTimeout(timeout).
		EnableDebugLog()

	return &Client{client: c}
}

func (c *Client) R() *req.Request {
	return c.client.R()
}

func (c *Client) SetProxy(proxyURL string) {
	c.client.SetProxyURL(proxyURL)
}
