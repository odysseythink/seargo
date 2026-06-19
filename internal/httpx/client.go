package httpx

import (
	"context"
	"fmt"
	"time"

	"github.com/seargo/seargo/internal/logger"
)

// Client is a network-aware HTTP client bound to a Registry.
type Client struct {
	registry       *Registry
	networkName    string
	engineName     string
	defaultUA      string
	defaultTimeout time.Duration
}

// NewClient creates a Client bound to the given Registry. If registry is nil,
// it panics (startup error — fail fast).
func NewClient(registry *Registry, networkName, engineName string, defaultUA string, defaultTimeout time.Duration) *Client {
	if registry == nil {
		panic("httpx.NewClient: registry must not be nil")
	}
	return &Client{
		registry:       registry,
		networkName:    networkName,
		engineName:     engineName,
		defaultUA:      defaultUA,
		defaultTimeout: defaultTimeout,
	}
}

// R returns a new RequestBuilder for constructing and executing a request.
func (c *Client) R() *RequestBuilder {
	return &RequestBuilder{
		client:      c,
		queryParams: make(map[string]string),
		headers:     make(map[string]string),
		formData:    make(map[string]string),
	}
}

// SetProxy is a deprecated noop. Proxy configuration is managed by the Network.
func (c *Client) SetProxy(proxyURL string) {
	logger.Warn("Client.SetProxy is deprecated; proxy configuration is managed by Network", "engine", c.engineName)
}

// WithNetwork returns a copy of the Client bound to a different named Network.
func (c *Client) WithNetwork(name string) *Client {
	return &Client{
		registry:       c.registry,
		networkName:    name,
		engineName:     c.engineName,
		defaultUA:      c.defaultUA,
		defaultTimeout: c.defaultTimeout,
	}
}

// RequestBuilder is a chainable HTTP request builder.
type RequestBuilder struct {
	client       *Client
	method       string
	url          string
	queryParams  map[string]string
	headers      map[string]string
	body         []byte
	formData     map[string]string
	timeout      time.Duration
	maxRedirects int
}

func (rb *RequestBuilder) SetQueryParam(k, v string) *RequestBuilder {
	rb.queryParams[k] = v
	return rb
}

func (rb *RequestBuilder) SetQueryParams(m map[string]string) *RequestBuilder {
	for k, v := range m {
		rb.queryParams[k] = v
	}
	return rb
}

func (rb *RequestBuilder) SetHeader(k, v string) *RequestBuilder {
	rb.headers[k] = v
	return rb
}

func (rb *RequestBuilder) SetBody(body []byte) *RequestBuilder {
	rb.body = body
	return rb
}

func (rb *RequestBuilder) SetFormData(m map[string]string) *RequestBuilder {
	for k, v := range m {
		rb.formData[k] = v
	}
	return rb
}

func (rb *RequestBuilder) SetTimeout(d time.Duration) *RequestBuilder {
	rb.timeout = d
	return rb
}

func (rb *RequestBuilder) SetMaxRedirects(n int) *RequestBuilder {
	rb.maxRedirects = n
	return rb
}

// Get executes a GET request.
func (rb *RequestBuilder) Get(url string) (*Response, error) {
	rb.method = "GET"
	rb.url = url
	return rb.Do(context.Background())
}

// Post executes a POST request.
func (rb *RequestBuilder) Post(url string) (*Response, error) {
	rb.method = "POST"
	rb.url = url
	return rb.Do(context.Background())
}

// Response holds an HTTP response.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
	URL        string
	Duration   time.Duration
}

// Do executes the built request. Stub — full implementation in Task 2.
func (rb *RequestBuilder) Do(ctx context.Context) (*Response, error) {
	return nil, fmt.Errorf("Do not implemented yet")
}
