package httpx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/odysseythink/mlog"
)

// ctxKeyHTTPDuration is an unexported context key for passing HTTP duration.
// The context value is a *time.Duration mutable holder, set by the scheduler
// before calling Do and written by Do after each HTTP round-trip.
type ctxKeyHTTPDuration struct{}

// ContextWithHTTPDuration returns a new context carrying a mutable HTTP duration
// holder that Do populates after each HTTP round-trip.
func ContextWithHTTPDuration(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyHTTPDuration{}, new(time.Duration))
}

// HTTPDurationFromContext extracts the HTTP request duration from context.
func HTTPDurationFromContext(ctx context.Context) (time.Duration, bool) {
	p, ok := ctx.Value(ctxKeyHTTPDuration{}).(*time.Duration)
	if !ok || p == nil {
		return 0, false
	}
	return *p, true
}

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
	mlog.Warning("Client.SetProxy is deprecated; proxy configuration is managed by Network", "engine", c.engineName)
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
	ctx          context.Context
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

func (rb *RequestBuilder) SetContext(ctx context.Context) *RequestBuilder {
	rb.ctx = ctx
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

// String returns the response body as a string.
func (r *Response) String() string {
	return string(r.Body)
}

// resolveNetwork resolves the network for this Client.
// Priority: explicit networkName → engineName → "default".
func (c *Client) resolveNetwork() (*Network, error) {
	if c.networkName != "" {
		n := c.registry.Get(c.networkName)
		if n == nil {
			return nil, fmt.Errorf("unknown network %q", c.networkName)
		}
		return n, nil
	}

	if c.engineName != "" {
		n := c.registry.Get(c.engineName)
		if n != nil {
			return n, nil
		}
	}

	n := c.registry.Get("default")
	if n == nil {
		return nil, fmt.Errorf("default network not found")
	}
	return n, nil
}

// chooseUserAgent selects a User-Agent string.
// Priority: network.UserAgent > defaultUA.
func chooseUserAgent(network *Network, defaultUA string, _ *UserAgentPool) string {
	if network != nil && network.UserAgent != "" {
		return network.UserAgent + network.UserAgentSuffix
	}
	return defaultUA
}

// Do executes the built request through the Client's network.
func (rb *RequestBuilder) Do(ctx context.Context) (*Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if rb.ctx != nil {
		ctx = rb.ctx
	}

	// 1a. Check request body size
	if len(rb.body) > maxRequestSize {
		return nil, fmt.Errorf("request body exceeds max size of %d bytes", maxRequestSize)
	}

	// 1b. Resolve network
	network, err := rb.client.resolveNetwork()
	if err != nil {
		return nil, err
	}

	// 2. Check HTTP disabled
	if !network.EnableHTTP && rb.url != "" {
		parsedScheme := parseScheme(rb.url)
		if parsedScheme == "http" {
			return nil, fmt.Errorf("HTTP protocol is disabled for network %q", network.Name)
		}
	}

	// 3. Determine timeout
	timeout := rb.effectiveTimeout(network)

	// 4. Determine max redirects
	maxR := rb.maxRedirects
	if maxR <= 0 {
		maxR = network.MaxRedirects
	}
	if maxR <= 0 {
		maxR = 30
	}

	// 5. Determine verify
	verify := rb.boolHeader("X-SearGo-Skip-Verify") == "" && network.Verify

	// 6. Select local address and proxy
	localAddr := network.nextLocalAddress()
	proxyDigest := network.nextProxyDigest()

	// 7. Get or create resty client from Network cache
	restyClient, err := network.GetClient(verify, maxR, localAddr, proxyDigest)
	if err != nil {
		return nil, fmt.Errorf("get network client: %w", err)
	}

	// 8. Set timeout on client and build resty request
	restyClient.SetTimeout(timeout)
	req := restyClient.R().
		SetContext(ctx).
		SetQueryParams(rb.queryParams).
		SetHeaders(rb.headers)

	if len(rb.body) > 0 {
		req.SetBody(rb.body)
	}
	if len(rb.formData) > 0 {
		req.SetFormData(rb.formData)
	}

	// 9. UA selection
	if _, hasUA := rb.headers["User-Agent"]; !hasUA {
		ua := chooseUserAgent(network, rb.client.defaultUA, nil)
		if ua != "" {
			req.SetHeader("User-Agent", ua)
		}
	}

	// 10. Execute
	start := time.Now()
	var restyResp *resty.Response
	switch rb.method {
	case "GET":
		restyResp, err = req.Get(rb.url)
	case "POST":
		restyResp, err = req.Post(rb.url)
	default:
		return nil, fmt.Errorf("unsupported method: %s", rb.method)
	}
	duration := time.Since(start)

	// Record HTTP round-trip duration in the context's mutable holder.
	if durationPtr, ok := ctx.Value(ctxKeyHTTPDuration{}).(*time.Duration); ok && durationPtr != nil {
		*durationPtr = duration
	}

	if err != nil {
		return nil, classifyTransportError(err)
	}

	// 11. Check response body size
	if len(restyResp.Body()) > maxResponseSize {
		return nil, fmt.Errorf("response body exceeds max size of %d bytes", maxResponseSize)
	}

	// 12. Build Response
	resp := &Response{
		StatusCode: restyResp.StatusCode(),
		Body:       restyResp.Body(),
		Headers:    restyResp.RawResponse.Header,
		URL:        restyResp.Request.URL,
		Duration:   duration,
	}

	// 12. HTTP error classification (stub)
	if err := raiseForHTTPError(resp); err != nil {
		return resp, err
	}

	// 13. Metrics and logging
	recordMetrics(network.Name, rb.client.engineName, resp.StatusCode, duration, nil)
	logResponse(rb.client.engineName, network.Name, rb.method, rb.url, resp.StatusCode, nil)

	return resp, nil
}

// effectiveTimeout returns the effective timeout: explicit > network > client default > 3s.
func (rb *RequestBuilder) effectiveTimeout(network *Network) time.Duration {
	if rb.timeout > 0 {
		return rb.timeout
	}
	if rb.client.defaultTimeout > 0 {
		return rb.client.defaultTimeout
	}
	if network != nil && network.Timeout > 0 {
		return rb.client.defaultTimeout
	}
	return 3 * time.Second
}

func (rb *RequestBuilder) boolHeader(key string) string {
	return rb.headers[key]
}

func parseScheme(rawURL string) string {
	for i := 0; i < len(rawURL); i++ {
		if rawURL[i] == ':' {
			return rawURL[:i]
		}
		if rawURL[i] == '/' {
			break
		}
	}
	return ""
}

// UserAgentPool holds OS and version data for generating random User-Agent strings.
type UserAgentPool struct {
	mu       sync.RWMutex
	OSes     []string `json:"os"`
	Template string   `json:"ua"`
	Versions []string `json:"versions"`
}
