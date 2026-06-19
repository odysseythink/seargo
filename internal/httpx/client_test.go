package httpx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/logger"
)

func init() {
	// Ensure logger is initialized for tests that call logger.Warn
	logger.Init("warn", "stderr")
}

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, err := NewRegistry(cfg)
	require.NoError(t, err)

	c := NewClient(reg, "", "test-engine", "TestUA/1.0", 10*time.Second)
	assert.NotNil(t, c)
	assert.NotNil(t, c.R())
}

func TestNewClient_NilRegistryPanics(t *testing.T) {
	assert.Panics(t, func() {
		NewClient(nil, "", "test", "", 0)
	}, "nil registry should panic at construction")
}

func TestClient_R_ReturnsRequestBuilder(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "test", "", 0)

	rb := c.R()
	assert.NotNil(t, rb)
	assert.Equal(t, c, rb.client)
}

func TestRequestBuilder_SetQueryParam(t *testing.T) {
	rb := &RequestBuilder{queryParams: make(map[string]string)}
	result := rb.SetQueryParam("q", "test")
	assert.Same(t, rb, result, "should return self for chaining")
	assert.Equal(t, "test", rb.queryParams["q"])
}

func TestRequestBuilder_SetHeader(t *testing.T) {
	rb := &RequestBuilder{headers: make(map[string]string)}
	rb.SetHeader("X-Custom", "value")
	assert.Equal(t, "value", rb.headers["X-Custom"])
}

func TestRequestBuilder_SetTimeout(t *testing.T) {
	rb := &RequestBuilder{}
	rb.SetTimeout(5 * time.Second)
	assert.Equal(t, 5*time.Second, rb.timeout)
}

func TestRequestBuilder_SetMaxRedirects(t *testing.T) {
	rb := &RequestBuilder{}
	rb.SetMaxRedirects(10)
	assert.Equal(t, 10, rb.maxRedirects)
}

func TestClient_WithNetwork(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "engine-a", "", 0)
	c2 := c.WithNetwork("ipv4")
	assert.NotSame(t, c, c2)
	assert.Equal(t, "ipv4", c2.networkName)
	assert.Equal(t, c.registry, c2.registry)
	assert.Equal(t, c.engineName, c2.engineName, "engineName should be preserved")
}

func TestResolveNetwork_ExplicitNetwork(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "ipv4", "test", "", 0)
	n, err := c.resolveNetwork()
	assert.NoError(t, err)
	assert.Equal(t, "ipv4", n.Name)
}

func TestResolveNetwork_EngineFallback(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{
			{Name: "google", Engine: "google", Timeout: 5.0},
		},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "google", "", 0)
	n, err := c.resolveNetwork()
	assert.NoError(t, err)
	assert.Equal(t, "google", n.Name)
}

func TestResolveNetwork_DefaultFallback(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "unknown-engine", "", 0)
	n, err := c.resolveNetwork()
	assert.NoError(t, err)
	assert.Equal(t, "default", n.Name)
}

func TestResolveNetwork_UnknownExplicitNetwork(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "missing", "", "", 0)
	_, err := c.resolveNetwork()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown network")
}

func TestChooseUserAgent_Priority(t *testing.T) {
	n := &Network{UserAgent: "NetworkUA/1.0"}
	ua := chooseUserAgent(n, "DefaultUA/1.0", nil)
	assert.Equal(t, "NetworkUA/1.0", ua)

	n2 := &Network{UserAgent: ""}
	ua2 := chooseUserAgent(n2, "DefaultUA/1.0", nil)
	assert.Equal(t, "DefaultUA/1.0", ua2)
}

func TestDo_HTTPDisabled(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      false,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "test", "", 0)
	_, err := c.R().Get("http://example.com/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP protocol is disabled")
}

func TestDo_TimeoutDefaults(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "test", "", 15*time.Second)
	rb := c.R()
	timeout := rb.effectiveTimeout(reg.Get("default"))
	assert.Equal(t, 15*time.Second, timeout, "should use client defaultTimeout")
}

func TestDo_TimeoutOverride(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "test", "", 15*time.Second)
	rb := c.R().SetTimeout(2 * time.Second)
	timeout := rb.effectiveTimeout(reg.Get("default"))
	assert.Equal(t, 2*time.Second, timeout, "explicit SetTimeout should override default")
}

func TestClient_SetProxy_DeprecatedNoop(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "test", "", 0)
	c.SetProxy("http://proxy:8080")
}

func TestDo_ContextCancelled(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "test", "", 0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	_, err := c.R().Do(ctx)
	assert.Error(t, err)
}

func TestDo_GET_Integration(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "test", "", 5*time.Second)

	rb := c.R().
		SetQueryParam("q", "test").
		SetHeader("Accept", "text/html").
		SetTimeout(2 * time.Second)

	assert.NotNil(t, rb)
	assert.Equal(t, "test", rb.queryParams["q"])
	assert.Equal(t, "text/html", rb.headers["Accept"])
	assert.Equal(t, 2*time.Second, rb.timeout)

	// The actual HTTP request will fail (no real server), but the builder
	// and resolveNetwork path are verified.
	_, err := rb.Get("http://127.0.0.1:1/nonexistent")
	assert.Error(t, err) // connection refused or timeout
}

func TestDo_POST_Builder(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	reg, _ := NewRegistry(cfg)
	c := NewClient(reg, "", "test", "", 0)

	rb := c.R().SetBody([]byte(`{"key":"value"}`))
	assert.Equal(t, []byte(`{"key":"value"}`), rb.body)

	_, err := rb.Post("http://127.0.0.1:1/nonexistent")
	assert.Error(t, err) // connection refused
}
