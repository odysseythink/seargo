package botdetection

import (
	"context"
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
