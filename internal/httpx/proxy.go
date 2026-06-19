package httpx

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// ProxyURL holds a parsed proxy configuration.
type ProxyURL struct {
	Scheme   string
	Host     string
	Port     int
	Username string
	Password string
}

func (p ProxyURL) String() string {
	u := url.URL{
		Scheme: p.Scheme,
		Host:   netJoinHostPort(p.Host, p.Port),
	}
	if p.Username != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}
	return u.String()
}

func netJoinHostPort(host string, port int) string {
	if port == 0 {
		return host
	}
	return host + ":" + strconv.Itoa(port)
}

func parseProxyURL(raw string) (ProxyURL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ProxyURL{}, fmt.Errorf("empty proxy URL")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ProxyURL{}, fmt.Errorf("parse proxy URL %q: %w", raw, err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme == "" {
		return ProxyURL{}, fmt.Errorf("proxy URL must have a scheme: %q", raw)
	}

	host := u.Hostname()
	if host == "" {
		return ProxyURL{}, fmt.Errorf("proxy URL must have a host: %q", raw)
	}

	port := portOrDefault(u, scheme)

	pu := ProxyURL{
		Scheme: scheme,
		Host:   host,
		Port:   port,
	}
	if u.User != nil {
		pu.Username = u.User.Username()
		pu.Password, _ = u.User.Password()
	}
	return pu, nil
}

func portOrDefault(u *url.URL, scheme string) int {
	if p := u.Port(); p != "" {
		port, err := strconv.Atoi(p)
		if err == nil {
			return port
		}
	}
	switch scheme {
	case "http":
		return 80
	case "https":
		return 443
	case "socks4", "socks5", "socks5h":
		return 1080
	default:
		return 0
	}
}

const allPattern = "all://"

// ProxySet holds lists of proxy URLs keyed by URL scheme pattern.
type ProxySet struct {
	byPattern map[string][]ProxyURL
	indices   map[string]int
	mu        *sync.Mutex
}

func (ps *ProxySet) Next() map[string]ProxyURL {
	if len(ps.byPattern) == 0 {
		return nil
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	result := make(map[string]ProxyURL, len(ps.byPattern))
	for pattern, list := range ps.byPattern {
		if len(list) == 0 {
			continue
		}
		idx := ps.indices[pattern]
		result[pattern] = list[idx%len(list)]
		ps.indices[pattern] = idx + 1
	}
	return result
}

var schemeNormalization = map[string]string{
	"http":    "http://",
	"https":   "https://",
	"socks4":  "socks4://",
	"socks5":  "socks5://",
	"socks5h": "socks5h://",
	"http:":   "http://",
	"https:":  "https://",
	"socks4:": "socks4://",
	"socks5:": "socks5://",
}

func normalizePattern(pattern string) string {
	if m, ok := schemeNormalization[pattern]; ok {
		return m
	}
	if strings.HasSuffix(pattern, "://") {
		return pattern
	}
	return pattern + "://"
}

func parseProxies(input interface{}) (ProxySet, error) {
	ps := ProxySet{
		byPattern: make(map[string][]ProxyURL),
		indices:   make(map[string]int),
		mu:        &sync.Mutex{},
	}

	if input == nil {
		return ps, nil
	}

	switch v := input.(type) {
	case string:
		u, err := parseProxyURL(v)
		if err != nil {
			return ps, err
		}
		ps.byPattern[allPattern] = []ProxyURL{u}

	case map[string]interface{}:
		for pattern, value := range v {
			normalized := normalizePattern(pattern)
			urls, err := parseProxyList(value)
			if err != nil {
				return ps, fmt.Errorf("proxy pattern %q: %w", pattern, err)
			}
			ps.byPattern[normalized] = urls
		}

	default:
		return ps, fmt.Errorf("unsupported proxies type: %T", input)
	}

	return ps, nil
}

func parseProxyList(value interface{}) ([]ProxyURL, error) {
	switch v := value.(type) {
	case string:
		u, err := parseProxyURL(v)
		if err != nil {
			return nil, err
		}
		return []ProxyURL{u}, nil
	case []interface{}:
		urls := make([]ProxyURL, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("proxy list element must be a string, got %T", item)
			}
			u, err := parseProxyURL(s)
			if err != nil {
				return nil, err
			}
			urls = append(urls, u)
		}
		return urls, nil
	default:
		return nil, fmt.Errorf("proxy value must be string or list, got %T", value)
	}
}

// Peek returns the currently-selected proxies without advancing indices.
func (ps *ProxySet) Peek() map[string]ProxyURL {
	if len(ps.byPattern) == 0 {
		return nil
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	result := make(map[string]ProxyURL, len(ps.byPattern))
	for pattern, list := range ps.byPattern {
		if len(list) == 0 {
			continue
		}
		idx := ps.indices[pattern]
		result[pattern] = list[idx%len(list)]
	}
	return result
}

func (ps *ProxySet) Len() int {
	total := 0
	for _, list := range ps.byPattern {
		total += len(list)
	}
	return total
}
