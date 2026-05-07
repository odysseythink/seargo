package httpx

import (
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

type Client struct {
	client *resty.Client
}

func New(userAgent string, timeout time.Duration) *Client {
	if userAgent == "" {
		userAgent = userAgents[rand.Intn(len(userAgents))]
	}
	c := resty.New().
		SetHeader("User-Agent", userAgent).
		SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8").
		SetHeader("Accept-Language", "en-US,en;q=0.5").
		SetHeader("Accept-Encoding", "gzip, deflate, br").
		SetHeader("DNT", "1").
		SetHeader("Connection", "keep-alive").
		SetTimeout(timeout).
		SetRetryCount(2).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)

	return &Client{client: c}
}

func (c *Client) R() *resty.Request {
	return c.client.R()
}

func (c *Client) SetProxy(proxyURL string) {
	c.client.SetProxy(proxyURL)
}

func RandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}
