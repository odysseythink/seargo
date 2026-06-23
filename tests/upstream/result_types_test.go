//go:build upstream

package upstream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResultTypes_SuggestionsAndCorrections(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name  string
		query string
	}{
		{"suggestion", "golang"},
		{"correction", "golng"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("suggest/"+tc.name, tc.query, SearchParams{})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
		})
	}
}

func TestResultTypes_Answers(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name  string
		query string
	}{
		{"answer_random", "random"},
		{"answer_statistics", "statistics 1 2 3"},
		{"answer_weather", "weather in Beijing"},
		{"answer_currency", "100 USD to CNY"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("answer/"+tc.name, tc.query, SearchParams{})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Answers, "%s: expected at least one answer", tc.name)
		})
	}
}

func TestResultTypes_Infoboxes(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name  string
		query string
	}{
		{"infobox_person", "Albert Einstein"},
		{"infobox_city",   "Beijing"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("infobox/"+tc.name, tc.query, SearchParams{})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Infoboxes, "%s: expected at least one infobox", tc.name)
		})
	}
}

func TestResultTypes_ImageAndVideo(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name     string
		query    string
		category string
		template string
	}{
		{"images", "golang", "images", "images.html"},
		{"videos", "golang", "videos", "videos.html"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("media/"+tc.name, tc.query, SearchParams{Category: tc.category})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Results, "%s: expected results", tc.name)

			var found bool
			for _, r := range report.Results {
				if r.Template == tc.template {
					found = true
					break
				}
			}
			require.True(t, found, "%s: expected at least one result with template %q", tc.name, tc.template)
		})
	}
}

func TestResultTypes_OtherKinds(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name     string
		query    string
		category string
		template string
	}{
		{"files",    "golang pdf", "files",   "files.html"},
		{"code",     "golang",     "it",      "code.html"},
		{"science",  "golang",     "science", "paper.html"},
		{"keyvalue", "golang",     "general", "keyvalue.html"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("kind/"+tc.name, tc.query, SearchParams{Category: tc.category})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Results, "%s: expected results", tc.name)
			for _, r := range report.Results {
				require.Equal(t, tc.template, r.Template, "%s: template mismatch", tc.name)
			}
		})
	}
}
