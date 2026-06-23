package upstream

import (
	"context"
	"os"
	"testing"
	"time"
)

// Config holds the parity-test runtime configuration.
type Config struct {
	SearGoBaseURL   string
	UpstreamBaseURL string
	Timeout         time.Duration
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
	cfg := &Config{
		SearGoBaseURL:   getenv("SEARGO_BASE_URL", "http://127.0.0.1:8080"),
		UpstreamBaseURL: getenv("UPSTREAM_BASE_URL", "http://127.0.0.1:8081"),
		Timeout:         mustDuration(getenv("UPSTREAM_TIMEOUT", "60s")),
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 60 * time.Second
	}
	return d
}

// Harness provides shared test setup for parity tests.
type Harness struct {
	T      *testing.T
	Client *Client
	Config *Config
	gaps   GapRules
}

// NewHarness creates a new test harness.
func NewHarness(t *testing.T) *Harness {
	t.Helper()
	cfg := LoadConfig()
	h := &Harness{
		T:      t,
		Client: NewClient(cfg),
		Config: cfg,
	}
	if rules, err := LoadGaps("tests/upstream/gaps.json"); err == nil {
		h.gaps = rules
	}
	return h
}

// RunCase calls both endpoints, diffs the responses, and returns the report.
func (h *Harness) RunCase(name, query string, params SearchParams) Report {
	h.T.Helper()
	ctx := context.Background()

	up, upErr := h.Client.SearchUpstream(ctx, query, params)
	if upErr != nil {
		h.T.Fatalf("upstream error for %s: %v", name, upErr)
	}
	sg, sgErr := h.Client.SearchSearGo(ctx, query, params)
	if sgErr != nil {
		h.T.Fatalf("seargo error for %s: %v", name, sgErr)
	}

	nup := NormalizeUpstream(up)
	nsg := NormalizeSearGo(sg)

	report := h.applyGaps(Report{
		Name:                name,
		Query:               query,
		Results:             nsg.Results,
		Answers:             nsg.Answers,
		Infoboxes:           nsg.Infoboxes,
		UnresponsiveEngines: nup.UnresponsiveEngines,
		FailedEngines:       nsg.UnresponsiveEngines,
		RedirectURL:         sg.RedirectURL,
		Mismatches:          Diff(name, nup, nsg),
	})
	GlobalReports.Record(report)
	return report
}

// RequireBothReachable skips the test if either endpoint is unreachable.
// Must be called at the test function level, before any t.Run subtests.
func (h *Harness) RequireBothReachable() {
	h.T.Helper()
	h.RequireUpstream()
	h.RequireSearGo()
}

// RequireUpstream skips the test if the upstream server is unreachable.
func (h *Harness) RequireUpstream() {
	h.T.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := h.Client.WaitForReady(ctx, h.Config.UpstreamBaseURL, "/")
	if err != nil {
		h.T.Skipf("upstream not reachable: %v", err)
	}
}

// RequireSearGo skips the test if the SearGo server is unreachable.
func (h *Harness) RequireSearGo() {
	h.T.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := h.Client.WaitForReady(ctx, h.Config.SearGoBaseURL, "/health")
	if err != nil {
		h.T.Skipf("seargo not reachable: %v", err)
	}
}

func (h *Harness) applyGaps(r Report) Report {
	if h.gaps == nil {
		return r
	}
	return h.gaps.Filter(r)
}

// WaitForUpstream blocks until the upstream server responds.
func (h *Harness) WaitForUpstream(ctx context.Context) error {
	return h.Client.WaitForReady(ctx, h.Config.UpstreamBaseURL, "/")
}

// WaitForSearGo blocks until the SearGo server responds.
func (h *Harness) WaitForSearGo(ctx context.Context) error {
	return h.Client.WaitForReady(ctx, h.Config.SearGoBaseURL, "/health")
}
