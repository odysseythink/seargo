package httpx

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func New(userAgent string, timeout time.Duration) *Client {
	c := resty.New().
		SetHeader("User-Agent", userAgent).
		SetTimeout(timeout)

	return &Client{client: c}
}

func (c *Client) R() *resty.Request {
	return c.client.R()
}

func (c *Client) SetProxy(proxyURL string) {
	c.client.SetProxy(proxyURL)
}
