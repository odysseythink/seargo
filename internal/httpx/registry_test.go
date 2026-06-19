package httpx

import (
	"testing"
	"time"

	"github.com/seargo/seargo/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetwork_GetClient_SameKeyReturnsSameClient(t *testing.T) {
	n := &Network{
		Name:                     "test",
		MaxConnections:           10,
		MaxKeepaliveConnections:  5,
		KeepaliveExpiry:          5 * time.Second,
		EnableHTTP2:              false,
		MaxRedirects:             5,
		clients:                  make(map[ClientKey]*restyClientRef),
	}

	c1, err := n.GetClient(true, 5, "", "")
	assert.NoError(t, err)
	c2, err := n.GetClient(true, 5, "", "")
	assert.NoError(t, err)
	assert.Same(t, c1, c2)
}

func TestNetwork_GetClient_DifferentVerifyCreatesNew(t *testing.T) {
	n := &Network{
		Name:                     "test",
		MaxConnections:           10,
		MaxKeepaliveConnections:  5,
		KeepaliveExpiry:          5 * time.Second,
		MaxRedirects:             5,
		clients:                  make(map[ClientKey]*restyClientRef),
	}

	c1, err := n.GetClient(true, 5, "", "")
	assert.NoError(t, err)
	c2, err := n.GetClient(false, 5, "", "")
	assert.NoError(t, err)
	assert.NotSame(t, c1, c2)
}

func TestNetwork_GetClient_DifferentLocalAddrCreatesNew(t *testing.T) {
	n := &Network{
		Name:                     "test",
		MaxConnections:           10,
		MaxKeepaliveConnections:  5,
		KeepaliveExpiry:          5 * time.Second,
		MaxRedirects:             5,
		LocalAddresses:           []string{"10.0.0.1", "10.0.0.2"},
		clients:                  make(map[ClientKey]*restyClientRef),
	}

	c1, err := n.GetClient(true, 5, "", "")
	assert.NoError(t, err)
	c2, err := n.GetClient(true, 5, "", "")
	assert.NoError(t, err)
	assert.NotSame(t, c1, c2, "different local address should produce different client")
}

func TestNetwork_GetClient_ProxyRoundRobin(t *testing.T) {
	ps, _ := parseProxies(map[string]interface{}{
		"all": []interface{}{"http://a:8080", "http://b:8080"},
	})
	n := &Network{
		Name:                     "test",
		MaxConnections:           10,
		MaxKeepaliveConnections:  5,
		KeepaliveExpiry:          5 * time.Second,
		MaxRedirects:             5,
		Proxies:                  ps,
		clients:                  make(map[ClientKey]*restyClientRef),
	}

	c1, err := n.GetClient(true, 5, "", "")
	assert.NoError(t, err)
	c2, err := n.GetClient(true, 5, "", "")
	assert.NoError(t, err)
	assert.NotSame(t, c1, c2, "proxy round-robin should produce different client")
}

func TestNetwork_Close(t *testing.T) {
	n := &Network{
		Name:                     "test",
		MaxConnections:           10,
		MaxKeepaliveConnections:  5,
		KeepaliveExpiry:          5 * time.Second,
		MaxRedirects:             5,
		clients:                  make(map[ClientKey]*restyClientRef),
	}

	_, err := n.GetClient(true, 5, "", "")
	assert.NoError(t, err)

	n.Close()

	_, err = n.GetClient(true, 5, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestNetwork_ClientKey_ProxyDigestStable(t *testing.T) {
	ps, _ := parseProxies("http://a:8080")
	n := &Network{
		Name:    "test",
		Proxies: ps,
	}

	digest1 := n.proxyDigest()
	digest2 := n.proxyDigest()
	assert.Equal(t, digest1, digest2, "same proxy set gives same digest")

	n2 := &Network{Name: "empty"}
	assert.Equal(t, "", n2.proxyDigest())
}

func TestRegistry_Initialize_CreatesDefault(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:    3.0,
			PoolConnections:   100,
			PoolMaxsize:       10,
			KeepaliveExpiry:   5.0,
			MaxRedirects:      30,
			EnableHTTP:        true,
			Retries:           0,
		},
		Engines: []config.EngineConfig{
			{Name: "google", Engine: "google", Timeout: 10.0},
		},
	}

	r, err := NewRegistry(cfg)
	require.NoError(t, err)
	assert.NotNil(t, r.Get("default"))
	assert.NotNil(t, r.Get("ipv4"))
	assert.NotNil(t, r.Get("ipv6"))
	assert.NotNil(t, r.Get("google"))
	assert.NotNil(t, r.Get("image_proxy"))
}

func TestRegistry_Initialize_CustomNetwork(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
			Networks: map[string]config.OutgoingNetworkOverride{
				"tor": {
					UsingTorProxy: boolPtr(true),
				},
			},
		},
		Engines: []config.EngineConfig{},
	}

	r, err := NewRegistry(cfg)
	require.NoError(t, err)
	assert.NotNil(t, r.Get("tor"))
}

func TestRegistry_Initialize_DuplicateBuiltinFails(t *testing.T) {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
			Networks: map[string]config.OutgoingNetworkOverride{
				"default": {},
			},
		},
		Engines: []config.EngineConfig{},
	}

	_, err := NewRegistry(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts")
}

func TestRegistry_Initialize_EngineNetwork(t *testing.T) {
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
			{Name: "bing", Engine: "bing", Timeout: 5.0},
		},
	}

	r, err := NewRegistry(cfg)
	require.NoError(t, err)
	bingNet := r.Get("bing")
	assert.NotNil(t, bingNet)
	assert.Equal(t, 5*time.Second, bingNet.Timeout)
}

func TestRegistry_Get_Missing(t *testing.T) {
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

	r, _ := NewRegistry(cfg)
	assert.Nil(t, r.Get("nonexistent"))
}

func TestRegistry_Close(t *testing.T) {
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

	r, _ := NewRegistry(cfg)
	assert.NoError(t, r.Close())
}

func TestRegistry_Reload_ReplacesNetworks(t *testing.T) {
	cfg1 := &config.Config{
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

	r, err := NewRegistry(cfg1)
	require.NoError(t, err)
	origDefault := r.Get("default")
	assert.NotNil(t, origDefault)

	cfg2 := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  10.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}

	err = r.Reload(cfg2)
	require.NoError(t, err)

	newDefault := r.Get("default")
	assert.NotNil(t, newDefault)
	assert.NotSame(t, origDefault, newDefault)
	assert.Equal(t, 10*time.Second, newDefault.Timeout)
}

func TestRegistry_Reload_FailureKeepsOld(t *testing.T) {
	cfg1 := &config.Config{
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

	r, err := NewRegistry(cfg1)
	require.NoError(t, err)
	origDefault := r.Get("default")

	cfg2 := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  10.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
			Networks: map[string]config.OutgoingNetworkOverride{
				"default": {},
			},
		},
		Engines: []config.EngineConfig{},
	}

	err = r.Reload(cfg2)
	assert.Error(t, err)

	stillDefault := r.Get("default")
	assert.NotNil(t, stillDefault)
	assert.Same(t, origDefault, stillDefault)
}

func TestRegistry_Reload_AddsNewEngine(t *testing.T) {
	cfg1 := &config.Config{
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

	r, err := NewRegistry(cfg1)
	require.NoError(t, err)
	assert.Nil(t, r.Get("google"))

	cfg2 := &config.Config{
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

	err = r.Reload(cfg2)
	require.NoError(t, err)
	assert.NotNil(t, r.Get("google"))
}

func boolPtr(b bool) *bool { return &b }
