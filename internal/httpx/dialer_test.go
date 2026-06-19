package httpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDialContext_HTTPProxy(t *testing.T) {
	pu := ProxyURL{Scheme: "http", Host: "proxy.example.com", Port: 8080}
	dialCtx, err := newDialContext(pu, "")
	assert.NoError(t, err)
	assert.Nil(t, dialCtx, "HTTP/HTTPS proxy should not produce a custom dial context (handled by transport.Proxy)")
}

func TestNewDialContext_SOCKS5(t *testing.T) {
	pu := ProxyURL{Scheme: "socks5", Host: "127.0.0.1", Port: 1080}
	dialCtx, err := newDialContext(pu, "")
	assert.NoError(t, err)
	assert.NotNil(t, dialCtx, "SOCKS5 proxy should produce a dial context")
}

func TestNewDialContext_SOCKS5H(t *testing.T) {
	pu := ProxyURL{Scheme: "socks5h", Host: "127.0.0.1", Port: 1080}
	dialCtx, err := newDialContext(pu, "")
	assert.NoError(t, err)
	assert.NotNil(t, dialCtx, "SOCKS5H proxy should produce a dial context")
}

func TestNewDialContext_InvalidScheme(t *testing.T) {
	pu := ProxyURL{Scheme: "ftp", Host: "proxy.example.com", Port: 21}
	_, err := newDialContext(pu, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported proxy scheme")
}

func TestNewRestyClient_BasicConfig(t *testing.T) {
	n := &Network{
		Name:                     "test",
		MaxConnections:           100,
		MaxKeepaliveConnections:  20,
		KeepaliveExpiry:          5 * time.Second,
		EnableHTTP2:              false,
		MaxRedirects:             10,
		clients:                  make(map[ClientKey]*restyClientRef),
	}

	rc, err := n.newRestyClient(true, 10, "", "")
	require.NoError(t, err)
	assert.NotNil(t, rc)

	transport := rc.GetClient().Transport
	assert.NotNil(t, transport)
}

func TestNewRestyClient_WithHTTPProxy(t *testing.T) {
	ps, _ := parseProxies("http://proxy:8080")
	n := &Network{
		Name:                     "test",
		MaxConnections:           100,
		MaxKeepaliveConnections:  20,
		KeepaliveExpiry:          5 * time.Second,
		MaxRedirects:             10,
		Proxies:                  ps,
		clients:                  make(map[ClientKey]*restyClientRef),
	}

	rc, err := n.newRestyClient(true, 10, "", n.proxyDigest())
	require.NoError(t, err)
	assert.NotNil(t, rc)
}
