package upstream

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHarness_GapSuppression(t *testing.T) {
	cfg := &Config{SearGoBaseURL: "http://localhost:1", UpstreamBaseURL: "http://localhost:1", Timeout: time.Second}
	h := &Harness{T: t, Config: cfg, gaps: GapRules{
		{Name: "gaptest", PathPrefix: "gaptest.query", Reason: "demo"},
	}}
	r := h.applyGaps(Report{Name: "gaptest", Mismatches: []Mismatch{{Path: "gaptest.query", Want: "a", Got: "b"}}})
	require.Empty(t, r.Mismatches)
	require.Len(t, r.Suppressed, 1)
}

func TestConfig_FromEnv(t *testing.T) {
	t.Setenv("SEARGO_BASE_URL", "http://127.0.0.1:9090")
	t.Setenv("UPSTREAM_BASE_URL", "http://127.0.0.1:9091")
	t.Setenv("UPSTREAM_TIMEOUT", "45s")

	cfg := LoadConfig()
	require.Equal(t, "http://127.0.0.1:9090", cfg.SearGoBaseURL)
	require.Equal(t, "http://127.0.0.1:9091", cfg.UpstreamBaseURL)
	require.Equal(t, 45*time.Second, cfg.Timeout)
}
