package httpx

import (
	"context"
	"fmt"
	"net"
	"time"

	netproxy "golang.org/x/net/proxy"
)

// newDialContext returns a custom DialContext for the given proxy URL.
// For HTTP/HTTPS proxies, it returns (nil, nil) because those are handled
// by http.Transport.Proxy. For SOCKS proxies, it returns a dialer that
// routes traffic through the SOCKS server.
func newDialContext(pu ProxyURL, localAddr string) (func(ctx context.Context, network, addr string) (net.Conn, error), error) {
	switch pu.Scheme {
	case "http", "https":
		// HTTP/HTTPS proxy is handled by http.Transport.Proxy
		return nil, nil
	case "socks4", "socks5", "socks5h":
		return socks5DialContext(pu, localAddr)
	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %q", pu.Scheme)
	}
}

// socks5DialContext creates a SOCKS5 dial context. When localAddr is non-empty,
// it uses a localDialer to bind to that address.
func socks5DialContext(pu ProxyURL, localAddr string) (func(ctx context.Context, network, addr string) (net.Conn, error), error) {
	auth := netproxy.Auth{}
	if pu.Username != "" {
		auth.User = pu.Username
		auth.Password = pu.Password
	}

	// Build the base dialer: either a local-address-bind dialer or Direct.
	var baseDialer netproxy.Dialer = netproxy.Direct
	if localAddr != "" {
		baseDialer = &localDialer{addr: localAddr}
	}

	socksDialer, err := netproxy.SOCKS5("tcp", net.JoinHostPort(pu.Host, fmt.Sprintf("%d", pu.Port)), &auth, baseDialer)
	if err != nil {
		return nil, fmt.Errorf("create SOCKS5 dialer: %w", err)
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return socksDialer.Dial(network, addr)
	}, nil
}

// localDialer implements netproxy.Dialer to bind outbound connections
// to a specific local IP address.
type localDialer struct {
	addr string
}

func (d *localDialer) Dial(network, addr string) (net.Conn, error) {
	var laddr net.Addr
	if network == "tcp" || network == "tcp4" || network == "tcp6" {
		tcpAddr, err := net.ResolveTCPAddr(network, net.JoinHostPort(d.addr, "0"))
		if err != nil {
			return nil, err
		}
		laddr = tcpAddr
	}

	dialer := net.Dialer{LocalAddr: laddr, Timeout: 30 * time.Second}
	return dialer.Dial(network, addr)
}
