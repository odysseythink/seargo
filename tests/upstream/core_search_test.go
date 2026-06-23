//go:build upstream

package upstream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreSearch_Categories(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name     string
		query    string
		category string
	}{
		{"general", "golang", "general"},
		{"images",  "golang", "images"},
		{"videos",  "golang", "videos"},
		{"news",    "golang", "news"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("category/"+tc.name, tc.query, SearchParams{Category: tc.category})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Results, "%s: expected at least one result", tc.name)
			for _, r := range report.Results {
				require.Equal(t, tc.category, r.Category, "%s: result category mismatch", tc.name)
			}
		})
	}
}

func TestCoreSearch_BangEngineSelection(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name   string
		query  string
		engine string
	}{
		{"bang_google",     "!g golang",   "google"},
		{"bang_bing",       "!b golang",   "bing"},
		{"bang_duckduckgo", "!ddg golang", "duckduckgo"},
		{"bang_wikipedia",  "!wp golang",  "wikipedia"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("bang/"+tc.name, tc.query, SearchParams{})
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
			require.NotEmpty(t, report.Results, "%s: expected results", tc.name)
			for _, r := range report.Results {
				require.Equal(t, tc.engine, r.Engine, "%s: engine mismatch", tc.name)
			}
		})
	}
}

func TestCoreSearch_Filters(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	cases := []struct {
		name   string
		query  string
		params SearchParams
	}{
		{"language_en",    ":en golang", SearchParams{}},
		{"language_zh",    ":zh-CN golang", SearchParams{}},
		{"safesearch_off", "golang", SearchParams{SafeSearch: 0}},
		{"safesearch_on",  "golang", SearchParams{SafeSearch: 1}},
		{"timerange_year", "golang", SearchParams{TimeRange: "year"}},
		{"timerange_day",  "golang", SearchParams{TimeRange: "day"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := h.RunCase("filter/"+tc.name, tc.query, tc.params)
			require.Empty(t, report.Mismatches, "%s: mismatches: %+v", tc.name, report.Mismatches)
		})
	}
}

func TestCoreSearch_Pagination(t *testing.T) {
	t.Parallel()
	h := NewHarness(t)
	h.RequireBothReachable()

	page1 := h.RunCase("pagination/page1", "golang", SearchParams{Page: 1, PageSize: 5})
	require.Empty(t, page1.Mismatches, "page1 mismatches: %+v", page1.Mismatches)
	require.Len(t, page1.Results, 5, "page1 should have exactly page_size results")

	page2 := h.RunCase("pagination/page2", "golang", SearchParams{Page: 2, PageSize: 5})
	require.Empty(t, page2.Mismatches, "page2 mismatches: %+v", page2.Mismatches)
	require.Len(t, page2.Results, 5, "page2 should have exactly page_size results")

	page1URLs := make(map[string]bool)
	for _, r := range page1.Results {
		page1URLs[r.URL] = true
	}
	for _, r := range page2.Results {
		require.False(t, page1URLs[r.URL], "page2 result %s already appeared on page1", r.URL)
	}
}
