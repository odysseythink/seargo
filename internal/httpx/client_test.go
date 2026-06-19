package httpx

import (
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
