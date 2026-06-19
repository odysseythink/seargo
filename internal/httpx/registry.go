package httpx

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// ClientKey uniquely identifies a resty client in the Network cache.
type ClientKey struct {
	Verify       bool
	MaxRedirects int
	LocalAddress string
	ProxyDigest  string
}

type restyClientRef struct {
	Client *resty.Client
}

// Network holds configuration for a named outbound network endpoint.
type Network struct {
	Name                     string
	EnableHTTP               bool
	Verify                   bool
	EnableHTTP2              bool
	MaxConnections           int
	MaxKeepaliveConnections  int
	KeepaliveExpiry          time.Duration
	LocalAddresses           []string
	Proxies                  ProxySet
	UsingTorProxy            bool
	MaxRedirects             int
	Retries                  int
	RetryOnHTTPError         interface{}
	UserAgent                string
	UserAgentSuffix          string
	Timeout                  time.Duration

	mu           sync.Mutex
	addressIndex int
	clients      map[ClientKey]*restyClientRef
	closed       bool
}

// GetClient returns a resty client for the given parameters, creating
// one if needed. The client is cached by ClientKey. When localAddr or
// proxyDigest are empty and the Network has multiple local addresses or
// proxies configured, they are auto-selected using round-robin.
func (n *Network) GetClient(verify bool, maxRedirects int, localAddr, proxyDigest string) (*resty.Client, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return nil, fmt.Errorf("network %q is closed", n.Name)
	}

	// Auto-select local address (round-robin) when not explicitly provided.
	if localAddr == "" && len(n.LocalAddresses) > 0 {
		idx := n.addressIndex
		n.addressIndex = (n.addressIndex + 1) % len(n.LocalAddresses)
		localAddr = n.LocalAddresses[idx]
	}

	// Auto-select proxy digest (round-robin) when not explicitly provided.
	if proxyDigest == "" && n.Proxies.Len() > 0 {
		selected := n.Proxies.Next()
		if len(selected) > 0 {
			proxyDigest = proxyDigestFromMap(selected)
		}
	}

	key := ClientKey{
		Verify:       verify,
		MaxRedirects: maxRedirects,
		LocalAddress: localAddr,
		ProxyDigest:  proxyDigest,
	}

	if ref, ok := n.clients[key]; ok && ref.Client != nil {
		return ref.Client, nil
	}

	rc, err := n.newRestyClient(verify, maxRedirects, localAddr, proxyDigest)
	if err != nil {
		return nil, err
	}

	n.clients[key] = &restyClientRef{Client: rc}
	return rc, nil
}

// nextLocalAddress returns the next local address for this network.
func (n *Network) nextLocalAddress() string {
	if len(n.LocalAddresses) == 0 {
		return ""
	}
	idx := n.addressIndex
	n.addressIndex = (n.addressIndex + 1) % len(n.LocalAddresses)
	return n.LocalAddresses[idx]
}

// nextProxyDigest returns a digest of the currently-selected proxies.
func (n *Network) nextProxyDigest() string {
	if n.Proxies.Len() == 0 {
		return ""
	}
	selected := n.Proxies.Next()
	if len(selected) == 0 {
		return ""
	}
	return proxyDigestFromMap(selected)
}

func proxyDigestFromMap(m map[string]ProxyURL) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{':'})
		h.Write([]byte(m[k].String()))
		h.Write([]byte{';'})
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// proxyDigest returns the current proxy digest without advancing.
func (n *Network) proxyDigest() string {
	if n.Proxies.Len() == 0 {
		return ""
	}
	selected := n.Proxies.Peek()
	if len(selected) == 0 {
		return ""
	}
	return proxyDigestFromMap(selected)
}

// Close closes all cached clients and marks the network as closed.
func (n *Network) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.closed = true
	for key, ref := range n.clients {
		if ref.Client != nil {
			ref.Client.GetClient().CloseIdleConnections()
		}
		delete(n.clients, key)
	}
	return nil
}

// newRestyClient constructs a resty.Client with a fully-configured http.Transport
// based on the Network settings.
func (n *Network) newRestyClient(verify bool, maxRedirects int, localAddr, proxyDigest string) (*resty.Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        n.MaxConnections,
		MaxIdleConnsPerHost: n.MaxKeepaliveConnections,
		IdleConnTimeout:     n.KeepaliveExpiry,
		ForceAttemptHTTP2:   n.EnableHTTP2,
	}

	if !verify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Proxy configuration
	if n.Proxies.Len() > 0 && proxyDigest != "" {
		selected := n.Proxies.Peek()
		pu, ok := selected[allPattern]
		if !ok {
			for _, v := range selected {
				pu = v
				break
			}
		}

		switch pu.Scheme {
		case "http", "https":
			proxyURLStr := pu.String()
			transport.Proxy = func(req *http.Request) (*url.URL, error) {
				u, err := url.Parse(proxyURLStr)
				if err != nil {
					return nil, err
				}
				return u, nil
			}
		case "socks4", "socks5", "socks5h":
			dialCtx, err := newDialContext(pu, localAddr)
			if err != nil {
				return nil, fmt.Errorf("SOCKS5 dialer: %w", err)
			}
			transport.DialContext = dialCtx
			transport.Proxy = nil
		}
	}

	// Local address binding
	if localAddr != "" && transport.DialContext == nil {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			tcpAddr, err := net.ResolveTCPAddr(network, net.JoinHostPort(localAddr, "0"))
			if err != nil {
				return nil, err
			}
			dialer := net.Dialer{LocalAddr: tcpAddr, Timeout: 30 * time.Second}
			return dialer.DialContext(ctx, network, addr)
		}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   0,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	rc := resty.NewWithClient(httpClient)
	return rc, nil
}
