package botdetection

import (
	"context"
	"net/http"
	"strings"
)

// DefaultUARegexps are the default UA regex patterns used when none are configured.
var DefaultUARegexps = []string{
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

// Detector runs the full probe set against a request.
type Detector struct {
	cfg    *Config
	probes []Probe
}

// NewDetector creates a Detector with all probes.
func NewDetector(cfg *Config, state State) *Detector {
	patterns := cfg.UserAgentPatterns
	if len(patterns) == 0 {
		patterns = DefaultUARegexps
	}
	probes := []Probe{
		&ipListProbe{},
		newUserAgentProbe(patterns),
		&acceptProbe{},
		&acceptEncodingProbe{},
		&acceptLanguageProbe{},
		&connectionProbe{},
		&secFetchProbe{},
	}
	if state != nil {
		probes = append(probes, newLinkTokenProbe(state))
	}
	return &Detector{cfg: cfg, probes: probes}
}

// Filter runs all probes and returns the first non-Allow decision.
func (d *Detector) Filter(ctx context.Context, req *http.Request, clientIP string) (Decision, string, error) {
	path := req.URL.Path

	if isExemptPath(path) {
		return Allow, "", nil
	}

	for _, probe := range d.probes {
		dec, err := probe.Filter(ctx, req, d.cfg, clientIP)
		if err != nil {
			return Block, probe.Name(), err
		}
		if dec != Allow {
			return dec, probe.Name(), nil
		}
	}

	return Allow, "", nil
}

func isExemptPath(path string) bool {
	if path == "/" || path == "/health" || path == "/metrics" || path == "/robots.txt" {
		return true
	}
	// API endpoints and static assets served by the embedded React frontend
	// should not be blocked by bot detection heuristics.
	if strings.HasPrefix(path, "/api/") {
		return true
	}
	// Client-side SPA routes — these serve index.html and should not be blocked.
	spaRoutes := []string{"/settings", "/stats", "/about", "/privacy", "/preferences"}
	for _, r := range spaRoutes {
		if path == r {
			return true
		}
	}
	return strings.HasPrefix(path, "/assets/") ||
		path == "/favicon.svg" ||
		path == "/icons.svg" ||
		strings.HasPrefix(path, "/locales/") ||
		path == "/searxng.png" ||
		path == "/empty_favicon.svg" ||
		path == "/img_load_error.svg" ||
		strings.HasSuffix(path, ".svg") && !strings.Contains(path, "/api/")
}
