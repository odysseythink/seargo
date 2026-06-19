package httpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProxyURL_HTTP(t *testing.T) {
	u, err := parseProxyURL("http://user:pass@proxy.example.com:8080")
	require.NoError(t, err)
	assert.Equal(t, "http", u.Scheme)
	assert.Equal(t, "proxy.example.com", u.Host)
	assert.Equal(t, 8080, u.Port)
	assert.Equal(t, "user", u.Username)
	assert.Equal(t, "pass", u.Password)
}

func TestParseProxyURL_SOCKS5(t *testing.T) {
	u, err := parseProxyURL("socks5://192.168.1.1:1080")
	require.NoError(t, err)
	assert.Equal(t, "socks5", u.Scheme)
	assert.Equal(t, "192.168.1.1", u.Host)
	assert.Equal(t, 1080, u.Port)
}

func TestParseProxyURL_Invalid(t *testing.T) {
	_, err := parseProxyURL("not a url")
	assert.Error(t, err)
}

func TestParseProxyURL_NoPort_Defaults(t *testing.T) {
	u, err := parseProxyURL("http://proxy.example.com")
	require.NoError(t, err)
	assert.Equal(t, 80, u.Port, "default HTTP port should be 80")
}

func TestParseProxyURL_SOCKS5_Defaults(t *testing.T) {
	u, err := parseProxyURL("socks5://proxy.example.com")
	require.NoError(t, err)
	assert.Equal(t, 1080, u.Port, "default SOCKS5 port should be 1080")
}

func TestNormalizePattern_Bare(t *testing.T) {
	assert.Equal(t, "socks5://", normalizePattern("socks5"))
	assert.Equal(t, "socks5h://", normalizePattern("socks5h"))
	assert.Equal(t, "http://", normalizePattern("http"))
	assert.Equal(t, "https://", normalizePattern("https"))
}

func TestNormalizePattern_AlreadyHasScheme(t *testing.T) {
	assert.Equal(t, "http://", normalizePattern("http://"))
	assert.Equal(t, "socks5://", normalizePattern("socks5://"))
	assert.Equal(t, "ftp://", normalizePattern("ftp://"))
}

func TestNormalizePattern_Colon(t *testing.T) {
	assert.Equal(t, "http://", normalizePattern("http:"))
	assert.Equal(t, "https://", normalizePattern("https:"))
}

func TestParseProxies_String(t *testing.T) {
	ps, err := parseProxies("http://proxy:8080")
	require.NoError(t, err)
	require.Len(t, ps.byPattern, 1)
	assert.Len(t, ps.byPattern["all://"], 1)
	assert.Equal(t, "http", ps.byPattern["all://"][0].Scheme)
}

func TestParseProxies_Dict(t *testing.T) {
	input := map[string]interface{}{
		"http":  "http://a:8080",
		"https": []interface{}{"http://b:8080", "http://c:8080"},
	}
	ps, err := parseProxies(input)
	require.NoError(t, err)
	assert.Len(t, ps.byPattern["http://"], 1)
	assert.Len(t, ps.byPattern["https://"], 2)
}

func TestParseProxies_AllPattern(t *testing.T) {
	ps, err := parseProxies("socks5://tor:9050")
	require.NoError(t, err)
	allList := ps.byPattern["all://"]
	require.Len(t, allList, 1)
	assert.Equal(t, "socks5", allList[0].Scheme)
}

func TestParseProxies_Nil(t *testing.T) {
	ps, err := parseProxies(nil)
	require.NoError(t, err)
	assert.Empty(t, ps.byPattern)
}

func TestProxySet_Next_RoundRobin(t *testing.T) {
	input := map[string]interface{}{
		"https": []interface{}{"http://a:8080", "http://b:8080"},
	}
	ps, _ := parseProxies(input)

	next := ps.Next()
	assert.Equal(t, "a", next["https://"].Host)

	next2 := ps.Next()
	assert.Equal(t, "b", next2["https://"].Host)

	next3 := ps.Next()
	assert.Equal(t, "a", next3["https://"].Host)
}

func TestProxySet_Next_Empty(t *testing.T) {
	ps := &ProxySet{byPattern: make(map[string][]ProxyURL)}
	assert.Empty(t, ps.Next())
}

func TestExpandLocalAddresses_Nil(t *testing.T) {
	addrs, err := expandLocalAddresses(nil)
	require.NoError(t, err)
	assert.Empty(t, addrs)
}

func TestExpandLocalAddresses_SingleIP(t *testing.T) {
	addrs, err := expandLocalAddresses("192.168.1.1")
	require.NoError(t, err)
	assert.Equal(t, []string{"192.168.1.1"}, addrs)
}

func TestExpandLocalAddresses_IPList(t *testing.T) {
	addrs, err := expandLocalAddresses([]interface{}{"10.0.0.1", "10.0.0.2"})
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, addrs)
}

func TestExpandLocalAddresses_CIDR(t *testing.T) {
	// /30 → 2 usable hosts: .1 and .2
	addrs, err := expandLocalAddresses("192.168.1.0/30")
	require.NoError(t, err)
	assert.Len(t, addrs, 2)
	assert.Equal(t, "192.168.1.1", addrs[0])
	assert.Equal(t, "192.168.1.2", addrs[1])
}

func TestExpandLocalAddresses_CIDRTooLarge(t *testing.T) {
	// /16 → 65534 hosts, should be rejected > maxSourceIPs (1024)
	_, err := expandLocalAddresses("10.0.0.0/16")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many source_ips")
}

func TestExpandLocalAddresses_InvalidIP(t *testing.T) {
	_, err := expandLocalAddresses("not-an-ip")
	assert.Error(t, err)
}

func TestExpandLocalAddresses_InvalidCIDR(t *testing.T) {
	_, err := expandLocalAddresses("10.0.0.0/99")
	assert.Error(t, err)
}

func TestExpandLocalAddresses_MixedCIDRAndIP(t *testing.T) {
	addrs, err := expandLocalAddresses([]interface{}{"10.0.0.1", "192.168.1.0/30"})
	require.NoError(t, err)
	assert.Len(t, addrs, 3) // 1 IP + 2 from /30
}

func TestExpandLocalAddresses_IPv6(t *testing.T) {
	addrs, err := expandLocalAddresses("::1")
	require.NoError(t, err)
	assert.Equal(t, []string{"::1"}, addrs)
}
