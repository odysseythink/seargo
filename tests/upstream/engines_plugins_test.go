//go:build upstream

package upstream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnginesPlugins_CategoryBang(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name     string
		query    string
		category string
	}{
		{"bang_general", "!general golang", "general"},
		{"bang_images",  "!images golang",  "images"},
		{"bang_videos",  "!videos golang",  "videos"},
		{"bang_news",    "!news golang",    "news"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("catbang/"+tc.name, tc.query, SearchParams{})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Results, "%s: expected results", tc.name)
			for _, r := range report.Results {
				require.Equal(t, tc.category, r.Category, "%s: category mismatch", tc.name)
			}
		})
	}
}

func TestEnginesPlugins_FailureReporting(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	report := h.RunCase("engine_failure", "xyzzy12345nonexistent", SearchParams{})
	require.Empty(t, report.Mismatches, "engine_failure: mismatches: %+v", report.Mismatches)

	upEngines := make(map[string]bool)
	for _, e := range report.UnresponsiveEngines {
		upEngines[e] = true
	}
	for _, e := range report.FailedEngines {
		require.True(t, upEngines[e], "upstream did not report %s as failed", e)
	}
}

func TestEnginesPlugins_CalculatorAndHash(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name  string
		query string
	}{
		{"plugin_calculator", "1 + 1"},
		{"plugin_hash_md5",   "md5 foo"},
		{"plugin_hash_sha256", "sha256 foo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("plugin/"+tc.name, tc.query, SearchParams{})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Answers, "%s: expected an answer", tc.name)
		})
	}
}

func TestEnginesPlugins_UnitAndTimeZone(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name  string
		query string
	}{
		{"plugin_unit_converter", "100 kg to lbs"},
		{"plugin_time_zone",      "time in Tokyo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("plugin/"+tc.name, tc.query, SearchParams{})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Answers, "%s: expected an answer", tc.name)
		})
	}
}

func TestEnginesPlugins_ExternalBang(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	upURL, err := h.Client.UpstreamExternalBangURL(t.Context(), "w", "golang")
	require.NoError(t, err)
	require.Contains(t, upURL, "wikipedia")

	report := h.RunCase("external_bang/wikipedia", "!!w golang", SearchParams{})
	require.Empty(t, report.Mismatches, "external bang mismatches: %+v", report.Mismatches)
	require.NotEmpty(t, report.RedirectURL, "expected a redirect URL")
	require.Equal(t, upURL, report.RedirectURL, "redirect URL mismatch")
}
