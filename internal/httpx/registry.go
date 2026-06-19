package httpx

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/seargo/seargo/internal/config"
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

// Registry holds all named outbound Networks.
type Registry struct {
	mu       sync.RWMutex
	networks map[string]*Network
	cfg      *config.Config
}

// NewRegistry creates a Registry and initializes all networks from config.
func NewRegistry(cfg *config.Config) (*Registry, error) {
	r := &Registry{
		networks: make(map[string]*Network),
		cfg:      cfg,
	}

	// 1. Default network
	defaultParams := buildParams(cfg.Outgoing, config.OutgoingNetworkOverride{})
	r.networks["default"] = newNetwork("default", defaultParams)

	// 2. Built-in ipv4 / ipv6
	ipv4Params := defaultParams
	ipv4Params.localAddrs = []string{"0.0.0.0"}
	r.networks["ipv4"] = newNetwork("ipv4", ipv4Params)

	ipv6Params := defaultParams
	ipv6Params.localAddrs = []string{"::"}
	r.networks["ipv6"] = newNetwork("ipv6", ipv6Params)

	// 3. Custom outgoing.networks
	for name, override := range cfg.Outgoing.Networks {
		if _, exists := r.networks[name]; exists {
			return nil, fmt.Errorf("network name %q conflicts with built-in network", name)
		}
		params := buildParams(cfg.Outgoing, override)
		r.networks[name] = newNetwork(name, params)
	}

	// 4. Engine networks
	for _, ec := range cfg.Engines {
		engineName := ec.Engine
		if engineName == "" {
			engineName = ec.Name
		}
		if engineName == "" {
			continue
		}
		params := defaultParams
		if ec.Timeout > 0 {
			params.timeout = time.Duration(ec.Timeout * float64(time.Second))
		}
		r.networks[engineName] = newNetwork(engineName, params)
	}

	// 5. image_proxy network
	if _, exists := r.networks["image_proxy"]; !exists {
		ipParams := defaultParams
		ipParams.enableHTTP2 = false
		r.networks["image_proxy"] = newNetwork("image_proxy", ipParams)
	}

	// 6. Tor validation
	for _, n := range r.networks {
		if n.UsingTorProxy {
			if err := n.checkTorProxy(); err != nil {
				return nil, fmt.Errorf("network %q is configured for Tor but check failed: %w", n.Name, err)
			}
		}
	}

	return r, nil
}

// Get returns the named Network or nil if not found.
func (r *Registry) Get(name string) *Network {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.networks[name]
}

// Names returns all registered network names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.networks))
	for name := range r.networks {
		names = append(names, name)
	}
	return names
}

// Reload rebuilds the Registry with a new config. If the new config is
// invalid, the old Registry is kept unchanged and an error is returned.
// On success, old Network clients are closed asynchronously.
func (r *Registry) Reload(newCfg *config.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Build new registry
	newRegistry := &Registry{
		networks: make(map[string]*Network),
		cfg:      newCfg,
	}

	defaultParams := buildParams(newCfg.Outgoing, config.OutgoingNetworkOverride{})
	newRegistry.networks["default"] = newNetwork("default", defaultParams)

	ipv4Params := defaultParams
	ipv4Params.localAddrs = []string{"0.0.0.0"}
	newRegistry.networks["ipv4"] = newNetwork("ipv4", ipv4Params)

	ipv6Params := defaultParams
	ipv6Params.localAddrs = []string{"::"}
	newRegistry.networks["ipv6"] = newNetwork("ipv6", ipv6Params)

	for name, override := range newCfg.Outgoing.Networks {
		if _, exists := newRegistry.networks[name]; exists {
			return fmt.Errorf("network name %q conflicts with built-in network", name)
		}
		params := buildParams(newCfg.Outgoing, override)
		newRegistry.networks[name] = newNetwork(name, params)
	}

	for _, ec := range newCfg.Engines {
		engineName := ec.Engine
		if engineName == "" {
			engineName = ec.Name
		}
		if engineName == "" {
			continue
		}
		params := defaultParams
		if ec.Timeout > 0 {
			params.timeout = time.Duration(ec.Timeout * float64(time.Second))
		}
		newRegistry.networks[engineName] = newNetwork(engineName, params)
	}

	if _, exists := newRegistry.networks["image_proxy"]; !exists {
		ipParams := defaultParams
		ipParams.enableHTTP2 = false
		newRegistry.networks["image_proxy"] = newNetwork("image_proxy", ipParams)
	}

	// Validate Tor
	for _, n := range newRegistry.networks {
		if n.UsingTorProxy {
			if err := n.checkTorProxy(); err != nil {
				return fmt.Errorf("network %q: %w", n.Name, err)
			}
		}
	}

	// Swap networks
	oldNetworks := r.networks
	r.networks = newRegistry.networks
	r.cfg = newCfg

	// Asynchronously close old network clients
	go func() {
		for _, n := range oldNetworks {
			n.Close()
		}
	}()

	return nil
}

// Close closes all networks and their clients.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []string
	for _, n := range r.networks {
		if err := n.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// networkParams is an internal, fully-resolved version of Network parameters
// used by buildParams to accumulate defaults and overrides.
type networkParams struct {
	enableHTTP              bool
	verify                  bool
	enableHTTP2             bool
	maxConnections          int
	maxKeepaliveConnections int
	keepaliveExpiry         time.Duration
	localAddrs              []string
	proxies                 ProxySet
	usingTorProxy           bool
	maxRedirects            int
	retries                 int
	retryOnHTTPError        interface{}
	userAgent               string
	userAgentSuffix         string
	timeout                 time.Duration
}

func buildParams(outgoing config.OutgoingConfig, override config.OutgoingNetworkOverride) networkParams {
	p := networkParams{
		enableHTTP:              true,
		verify:                  true,
		enableHTTP2:             outgoing.EnableHTTP2,
		maxConnections:          outgoing.PoolConnections,
		maxKeepaliveConnections: outgoing.PoolMaxsize,
		keepaliveExpiry:         time.Duration(outgoing.KeepaliveExpiry * float64(time.Second)),
		maxRedirects:            outgoing.MaxRedirects,
		retries:                 outgoing.Retries,
		retryOnHTTPError:        outgoing.RetryOnHTTPError,
		userAgent:               outgoing.UserAgent,
		userAgentSuffix:         outgoing.UserAgentSuffix,
		usingTorProxy:           outgoing.UsingTorProxy,
	}

	p.enableHTTP = outgoing.EnableHTTP

	if outgoing.MaxRedirects > 0 {
		p.maxRedirects = outgoing.MaxRedirects
	}
	if p.maxRedirects <= 0 {
		p.maxRedirects = 30
	}

	if outgoing.RequestTimeout > 0 {
		p.timeout = time.Duration(outgoing.RequestTimeout * float64(time.Second))
	}
	if p.timeout <= 0 {
		p.timeout = 3 * time.Second
	}

	// Apply overrides
	if override.EnableHTTP != nil {
		p.enableHTTP = *override.EnableHTTP
	}
	if override.Verify != nil {
		p.verify = *override.Verify
	}
	if override.EnableHTTP2 != nil {
		p.enableHTTP2 = *override.EnableHTTP2
	}
	if override.MaxConnections != nil {
		p.maxConnections = *override.MaxConnections
	}
	if override.MaxKeepaliveConnections != nil {
		p.maxKeepaliveConnections = *override.MaxKeepaliveConnections
	}
	if override.KeepaliveExpiry != nil {
		p.keepaliveExpiry = time.Duration(*override.KeepaliveExpiry * float64(time.Second))
	}
	if override.LocalAddresses != nil {
		addrs, err := expandLocalAddresses(override.LocalAddresses)
		if err == nil {
			p.localAddrs = addrs
		}
	}
	if override.Proxies != nil {
		ps, err := parseProxies(override.Proxies)
		if err == nil {
			p.proxies = ps
		}
	}
	if override.UsingTorProxy != nil {
		p.usingTorProxy = *override.UsingTorProxy
	}
	if override.MaxRedirects != nil {
		p.maxRedirects = *override.MaxRedirects
	}
	if override.Retries != nil {
		p.retries = *override.Retries
	}
	if override.RetryOnHTTPError != nil {
		p.retryOnHTTPError = override.RetryOnHTTPError
	}
	if override.UserAgent != "" {
		p.userAgent = override.UserAgent
	}
	if override.RequestTimeout != nil {
		p.timeout = time.Duration(*override.RequestTimeout * float64(time.Second))
	}
	if override.Timeout != nil {
		p.timeout = time.Duration(*override.Timeout * float64(time.Second))
	}

	// Apply outgoing-level proxies
	if outgoing.Proxies != nil {
		ps, err := parseProxies(outgoing.Proxies)
		if err == nil {
			p.proxies = ps
		}
	}
	if outgoing.SourceIPs != nil {
		addrs, err := expandLocalAddresses(outgoing.SourceIPs)
		if err == nil {
			p.localAddrs = addrs
		}
	}

	return p
}

func newNetwork(name string, p networkParams) *Network {
	maxConn := p.maxConnections
	if maxConn <= 0 {
		maxConn = 100
	}
	maxKeepalive := p.maxKeepaliveConnections
	if maxKeepalive <= 0 {
		maxKeepalive = 10
	}

	return &Network{
		Name:                     name,
		EnableHTTP:               p.enableHTTP,
		Verify:                   p.verify,
		EnableHTTP2:              p.enableHTTP2,
		MaxConnections:           maxConn,
		MaxKeepaliveConnections:  maxKeepalive,
		KeepaliveExpiry:          p.keepaliveExpiry,
		LocalAddresses:           p.localAddrs,
		Proxies:                  p.proxies,
		UsingTorProxy:            p.usingTorProxy,
		MaxRedirects:             p.maxRedirects,
		Retries:                  p.retries,
		RetryOnHTTPError:         p.retryOnHTTPError,
		UserAgent:                p.userAgent,
		UserAgentSuffix:          p.userAgentSuffix,
		Timeout:                  p.timeout,
		clients:                  make(map[ClientKey]*restyClientRef),
	}
}

// checkTorProxy verifies that this network's outbound IP is a Tor exit node.
// Uses https://check.torproject.org/api/ip endpoint.
func (n *Network) checkTorProxy() error {
	if !n.UsingTorProxy {
		return nil
	}

	if n.Proxies.Len() == 0 {
		return fmt.Errorf("using_tor_proxy is true but no proxy configured")
	}

	verify := n.Verify
	maxR := n.MaxRedirects
	if maxR <= 0 {
		maxR = 5
	}
	localAddr := n.nextLocalAddress()
	proxyDigest := n.nextProxyDigest()

	restyClient, err := n.GetClient(verify, maxR, localAddr, proxyDigest)
	if err != nil {
		return fmt.Errorf("create Tor check client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := restyClient.R().
		SetContext(ctx).
		Get("https://check.torproject.org/api/ip")
	if err != nil {
		return fmt.Errorf("Tor check request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("Tor check returned status %d", resp.StatusCode())
	}

	var result struct {
		IsTor bool   `json:"IsTor"`
		IP    string `json:"IP"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("Tor check response parse error: %w", err)
	}

	if !result.IsTor {
		return fmt.Errorf("Tor check failed: IP %s is not a Tor exit node", result.IP)
	}

	return nil
}
