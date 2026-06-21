package botdetection

import (
	"context"
	"net/http"
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
	return path == "/health" || path == "/metrics" || path == "/robots.txt"
}
