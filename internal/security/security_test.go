package security

import (
	"net"
	"net/http"
	"testing"
)

func TestHMACSigner_SignConsistency(t *testing.T) {
	s := NewHMACSigner("secret1")
	msg := []byte("http://example.com/a")
	sig1 := s.Sign(msg)
	sig2 := s.Sign(msg)
	if sig1 != sig2 {
		t.Fatalf("same message must produce same signature: %q vs %q", sig1, sig2)
	}
}

func TestHMACSigner_DifferentSecret(t *testing.T) {
	s1 := NewHMACSigner("secret1")
	s2 := NewHMACSigner("secret2")
	msg := []byte("hello")
	if s1.Sign(msg) == s2.Sign(msg) {
		t.Fatal("different secrets must produce different signatures")
	}
}

func TestHMACSigner_Verify(t *testing.T) {
	s := NewHMACSigner("mysecret")
	msg := []byte("http://example.com/a")
	sig := s.Sign(msg)

	// Valid signature
	if !s.Verify(msg, sig) {
		t.Fatal("Verify must return true for correct signature")
	}
	// Tampered signature — change last char
	tampered := sig[:len(sig)-1] + "x"
	if s.Verify(msg, tampered) {
		t.Fatal("Verify must return false for tampered signature")
	}
	// Empty signature
	if s.Verify(msg, "") {
		t.Fatal("Verify must return false for empty signature")
	}
}

func TestParseProxyList_Valid(t *testing.T) {
	pl, err := ParseProxyList([]string{"10.0.0.0/8", "127.0.0.0/8", "::1/128"})
	if err != nil {
		t.Fatal(err)
	}
	if len(pl) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(pl))
	}
	_, n1, _ := net.ParseCIDR("10.0.0.0/8")
	found := false
	for _, n := range pl {
		if n.String() == n1.String() {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("10.0.0.0/8 not found in parsed list")
	}
}

func TestParseProxyList_Invalid(t *testing.T) {
	_, err := ParseProxyList([]string{"not-a-cidr"})
	if err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
}

func TestIPExtractor_XFF(t *testing.T) {
	trusted, _ := ParseProxyList([]string{"10.0.0.0/8"})
	ext := NewIPExtractor(trusted)

	// Case A: X-Forwarded-For: untrusted,trusted → returns untrusted
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.1")
	ip := ext.ClientIP(req)
	if ip != "1.2.3.4" {
		t.Fatalf("Case A: got %q, want %q", ip, "1.2.3.4")
	}

	// Case B: X-Forwarded-For: all trusted → leftmost fallback
	req2, _ := http.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.2, 10.0.0.1")
	ip = ext.ClientIP(req2)
	if ip != "10.0.0.2" {
		t.Fatalf("Case B: got %q, want %q", ip, "10.0.0.2")
	}

	// Case C (adversarial): untrusted on right, trusted on left
	req3, _ := http.NewRequest("GET", "/", nil)
	req3.Header.Set("X-Forwarded-For", "10.0.0.1, 8.8.8.8")
	ip = ext.ClientIP(req3)
	if ip != "8.8.8.8" {
		t.Fatalf("Case C: got %q, want %q", ip, "8.8.8.8")
	}
}

func TestIPExtractor_RemoteAddr(t *testing.T) {
	trusted, _ := ParseProxyList([]string{})
	ext := NewIPExtractor(trusted)

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	ip := ext.ClientIP(req)
	if ip != "192.0.2.1" {
		t.Fatalf("got %q, want %q", ip, "192.0.2.1")
	}
}

func TestIPExtractor_XRealIP(t *testing.T) {
	trusted, _ := ParseProxyList([]string{})
	ext := NewIPExtractor(trusted)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "172.16.0.1")
	req.RemoteAddr = "127.0.0.1:9999"
	ip := ext.ClientIP(req)
	if ip != "172.16.0.1" {
		t.Fatalf("got %q, want %q", ip, "172.16.0.1")
	}
}

func TestIPExtractor_IPv6(t *testing.T) {
	trusted, _ := ParseProxyList([]string{"::1/128"})
	ext := NewIPExtractor(trusted)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "2001:db8::1, ::1")
	ip := ext.ClientIP(req)
	if ip != "2001:db8::1" {
		t.Fatalf("IPv6: got %q, want %q", ip, "2001:db8::1")
	}
}

func TestNetworkOf_IPv4(t *testing.T) {
	ip := net.ParseIP("1.2.3.4")
	n, err := NetworkOf(ip, 24, 48)
	if err != nil {
		t.Fatal(err)
	}
	if n.String() != "1.2.3.0/24" {
		t.Fatalf("got %q, want %q", n.String(), "1.2.3.0/24")
	}
}

func TestNetworkOf_IPv6(t *testing.T) {
	ip := net.ParseIP("2001:db8::1")
	n, err := NetworkOf(ip, 24, 48)
	if err != nil {
		t.Fatal(err)
	}
	if n.String() != "2001:db8::/48" {
		t.Fatalf("got %q, want %q", n.String(), "2001:db8::/48")
	}
}

func TestNetworkOf_InvalidPrefix(t *testing.T) {
	ip := net.ParseIP("1.2.3.4")
	n, err := NetworkOf(ip, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n.String() != "0.0.0.0/1" {
		t.Fatalf("got %q, want %q", n.String(), "0.0.0.0/1")
	}
}

func TestIsLinkLocal(t *testing.T) {
	if !IsLinkLocal(net.ParseIP("169.254.1.1")) {
		t.Fatal("169.254.1.1 should be link-local")
	}
	if !IsLinkLocal(net.ParseIP("fe80::1")) {
		t.Fatal("fe80::1 should be link-local")
	}
	if IsLinkLocal(net.ParseIP("192.168.1.1")) {
		t.Fatal("192.168.1.1 should NOT be link-local")
	}
}

func TestHashKey(t *testing.T) {
	h1 := HashKey("secret", "key1")
	h2 := HashKey("secret", "key1")
	if h1 != h2 {
		t.Fatalf("same inputs must hash same: %q vs %q", h1, h2)
	}
	h3 := HashKey("secret", "key2")
	if h1 == h3 {
		t.Fatal("different keys must hash differently")
	}
	h4 := HashKey("secret2", "key1")
	if h1 == h4 {
		t.Fatal("different secrets must hash differently")
	}
}
