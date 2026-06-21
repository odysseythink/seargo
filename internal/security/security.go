package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// HMACSigner signs and verifies messages with a single secret.
type HMACSigner interface {
	Sign(message []byte) string
	Verify(message []byte, signature string) bool
}

type hmacSigner struct {
	secret []byte
}

// NewHMACSigner creates an HMACSigner with the given secret.
func NewHMACSigner(secret string) HMACSigner {
	return &hmacSigner{secret: []byte(secret)}
}

func (s *hmacSigner) Sign(message []byte) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *hmacSigner) Verify(message []byte, signature string) bool {
	expected := s.Sign(message)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// IPExtractor extracts the real client IP, honoring trusted proxy CIDRs.
type IPExtractor interface {
	ClientIP(r *http.Request) string
}

type ipExtractor struct {
	trusted ProxyList
}

// NewIPExtractor creates an IPExtractor with the given trusted proxy list.
func NewIPExtractor(trusted ProxyList) IPExtractor {
	return &ipExtractor{trusted: trusted}
}

func (e *ipExtractor) ClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		addrs := splitCSV(xff)
		for i := len(addrs) - 1; i >= 0; i-- {
			addr := strings.TrimSpace(addrs[i])
			if addr == "" {
				continue
			}
			ip := net.ParseIP(addr)
			if ip == nil {
				continue
			}
			if !e.isTrusted(ip) {
				return ipString(ip)
			}
		}
		first := strings.TrimSpace(addrs[0])
		if ip := net.ParseIP(first); ip != nil {
			return ipString(ip)
		}
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		if ip := net.ParseIP(strings.TrimSpace(xri)); ip != nil {
			return ipString(ip)
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil {
		return ipString(ip)
	}
	return ""
}

func (e *ipExtractor) isTrusted(ip net.IP) bool {
	for _, network := range e.trusted {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func ipString(ip net.IP) string {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4.String()
	}
	return ip.String()
}

func splitCSV(s string) []string {
	return strings.Split(s, ",")
}

// ProxyList is a parsed list of trusted proxy CIDRs.
type ProxyList []*net.IPNet

// ParseProxyList parses a slice of CIDR strings into a ProxyList.
func ParseProxyList(raw []string) (ProxyList, error) {
	list := make(ProxyList, 0, len(raw))
	for _, cidr := range raw {
		cidr = strings.TrimSpace(cidr)
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
		}
		list = append(list, network)
	}
	return list, nil
}

// NetworkOf returns the network prefix for the given IP.
func NetworkOf(ip net.IP, v4Prefix, v6Prefix int) (*net.IPNet, error) {
	prefix := v6Prefix
	if ip.To4() != nil {
		prefix = v4Prefix
		prefix = clamp(prefix, 1, 32)
	} else {
		prefix = clamp(prefix, 1, 128)
	}
	_, network, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ip.String(), prefix))
	return network, err
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// IsLinkLocal reports whether ip is a link-local address.
func IsLinkLocal(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 169 && ip4[1] == 254
	}
	if len(ip) == net.IPv6len && ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
		return true
	}
	return false
}

// HashKey returns the lowercase hex HMAC-SHA256 of (secret, key).
func HashKey(secret, key string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}
