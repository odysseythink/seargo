package botdetection

import (
	"context"
	"net"
	"net/http"
	"testing"
)

func TestIPListProbe_Allow(t *testing.T) {
	cfg := &Config{}
	probe := &ipListProbe{}

	req, _ := http.NewRequest("GET", "/search?q=test", nil)
	dec, err := probe.Filter(context.Background(), req, cfg, "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("expected Allow, got %v", dec)
	}
}

func TestIPListProbe_BlockIP(t *testing.T) {
	cfg := &Config{
		IPLists: IPListsConfig{
			BlockIP: []string{"93.184.216.34"},
		},
	}
	probe := &ipListProbe{}

	req, _ := http.NewRequest("GET", "/", nil)
	dec, err := probe.Filter(context.Background(), req, cfg, "93.184.216.34")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("expected Block, got %v", dec)
	}
}

func TestIPListProbe_PassIP(t *testing.T) {
	cfg := &Config{
		IPLists: IPListsConfig{
			BlockIP: []string{"93.184.216.34"},
			PassIP:  []string{"8.8.8.8"},
		},
	}
	probe := &ipListProbe{}

	req, _ := http.NewRequest("GET", "/", nil)
	dec, err := probe.Filter(context.Background(), req, cfg, "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("expected Allow (pass_ip overrides block_ip), got %v", dec)
	}
}

func TestIPListProbe_InvalidIP(t *testing.T) {
	cfg := &Config{}
	probe := &ipListProbe{}

	req, _ := http.NewRequest("GET", "/", nil)
	dec, err := probe.Filter(context.Background(), req, cfg, "not-an-ip")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("invalid IP should Block, got %v", dec)
	}
}

var defaultUAPatterns = []string{
	`^$`,
	`(?i)curl/`,
	`(?i)wget/`,
	`(?i)python-requests/`,
	`(?i)scrapy`,
	`(?i)\bbot\b`,
	`(?i)\bcrawler\b`,
	`(?i)\bspider\b`,
	`(?i)\bheadless\b`,
}

func TestUserAgentProbe_Empty(t *testing.T) {
	cfg := &Config{}
	probe := newUserAgentProbe(defaultUAPatterns)

	req, _ := http.NewRequest("GET", "/", nil)
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("empty UA should Block, got %v", dec)
	}
}

func TestUserAgentProbe_BotDetected(t *testing.T) {
	cfg := &Config{}
	probe := newUserAgentProbe(defaultUAPatterns)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "curl/7.64.1")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("curl UA should Block, got %v", dec)
	}
}

func TestUserAgentProbe_BrowserAllowed(t *testing.T) {
	cfg := &Config{}
	probe := newUserAgentProbe(defaultUAPatterns)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("browser UA should Allow, got %v", dec)
	}
}

func TestUserAgentProbe_MustSurviveNormalBrowser(t *testing.T) {
	cfg := &Config{}
	probe := newUserAgentProbe(defaultUAPatterns)

	survivors := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148",
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0",
	}

	for _, ua := range survivors {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("User-Agent", ua)
		dec, err := probe.Filter(context.Background(), req, cfg, "")
		if err != nil {
			t.Fatalf("UA %q: unexpected error: %v", ua, err)
		}
		if dec != Allow {
			t.Fatalf("UA %q: expected Allow, got %v", ua, dec)
		}
	}
}

func TestLoadBotConfig(t *testing.T) {
	cfg, err := LoadConfig("../../configs/limiter.toml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("cfg is nil")
	}
	if cfg.IPv4Prefix != 32 {
		t.Fatalf("IPv4Prefix default: got %d, want 32", cfg.IPv4Prefix)
	}
}

func TestAcceptProbe_Allow(t *testing.T) {
	cfg := &Config{}
	probe := &acceptProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("accept text/html should Allow, got %v", dec)
	}
}

func TestAcceptProbe_Block(t *testing.T) {
	cfg := &Config{}
	probe := &acceptProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Accept", "application/json")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("accept application/json should Block, got %v", dec)
	}
}

func TestAcceptProbe_OnlyXHTML(t *testing.T) {
	cfg := &Config{}
	probe := &acceptProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Accept", "application/xhtml+xml")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("xhtml-only Accept should Block, got %v", dec)
	}
}

func TestAcceptEncodingProbe_Allow(t *testing.T) {
	cfg := &Config{}
	probe := &acceptEncodingProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("gzip encoding should Allow, got %v", dec)
	}
}

func TestAcceptEncodingProbe_Block(t *testing.T) {
	cfg := &Config{}
	probe := &acceptEncodingProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Accept-Encoding", "identity")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("identity encoding should Block, got %v", dec)
	}
}

func TestAcceptLanguageProbe_Allow(t *testing.T) {
	cfg := &Config{}
	probe := &acceptLanguageProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("Accept-Language should Allow, got %v", dec)
	}
}

func TestAcceptLanguageProbe_Block(t *testing.T) {
	cfg := &Config{}
	probe := &acceptLanguageProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("missing Accept-Language should Block, got %v", dec)
	}
}

func TestConnectionProbe_BlockClose(t *testing.T) {
	cfg := &Config{}
	probe := &connectionProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Connection", "close")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Block {
		t.Fatalf("Connection: close should Block, got %v", dec)
	}
}

func TestConnectionProbe_AllowKeepAlive(t *testing.T) {
	cfg := &Config{}
	probe := &connectionProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Connection", "keep-alive")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("Connection: keep-alive should Allow, got %v", dec)
	}
}

func TestSecFetchProbe_Allow(t *testing.T) {
	cfg := &Config{}
	probe := &secFetchProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("X-Forwarded-Proto", "https")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("Sec-Fetch browser headers should Allow, got %v", dec)
	}
}

func TestSecFetchProbe_Redirect(t *testing.T) {
	cfg := &Config{}
	probe := &secFetchProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Redirect {
		t.Fatalf("no-cors mode should Redirect, got %v", dec)
	}
}

func TestSecFetchProbe_PlainHTTP_Allow(t *testing.T) {
	cfg := &Config{}
	probe := &secFetchProbe{}

	req, _ := http.NewRequest("GET", "/search", nil)
	dec, err := probe.Filter(context.Background(), req, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("plain HTTP should Allow (can't enforce), got %v", dec)
	}
}

type mockState struct {
	suspicious bool
}

func (m *mockState) IsSuspicious(ctx context.Context, network *net.IPNet, acceptLanguage, userAgent string) (bool, error) {
	return m.suspicious, nil
}

func TestLinkTokenProbe_Suspicious(t *testing.T) {
	cfg := &Config{
		IPLimit: IPLimitConfig{LinkToken: true},
	}
	state := &mockState{suspicious: true}
	probe := newLinkTokenProbe(state)

	req, _ := http.NewRequest("GET", "/search", nil)
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	dec, err := probe.Filter(context.Background(), req, cfg, "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Redirect {
		t.Fatalf("suspicious network should Redirect, got %v", dec)
	}
}

func TestLinkTokenProbe_Disabled(t *testing.T) {
	cfg := &Config{
		IPLimit: IPLimitConfig{LinkToken: false},
	}
	state := &mockState{suspicious: true}
	probe := newLinkTokenProbe(state)

	req, _ := http.NewRequest("GET", "/search", nil)
	dec, err := probe.Filter(context.Background(), req, cfg, "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("disabled link_token should Allow, got %v", dec)
	}
}

func TestDetector_ExemptPaths(t *testing.T) {
	cfg := &Config{}
	det := NewDetector(cfg, nil)

	req, _ := http.NewRequest("GET", "/health", nil)
	dec, reason, err := det.Filter(context.Background(), req, "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("/health should be exempt, got %v reason=%q", dec, reason)
	}
}

func TestDetector_RunProbes(t *testing.T) {
	cfg := &Config{
		UserAgentPatterns: []string{},
	}
	det := NewDetector(cfg, nil)

	req, _ := http.NewRequest("GET", "/search?q=test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Connection", "keep-alive")

	dec, reason, err := det.Filter(context.Background(), req, "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if dec != Allow {
		t.Fatalf("normal browser should Allow, got %v reason=%q", dec, reason)
	}
	if reason != "" {
		t.Fatalf("reason should be empty for Allow, got %q", reason)
	}
}
