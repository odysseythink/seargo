package upstream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGaps_Filter(t *testing.T) {
	gaps := GapRules{
		{Name: "suggest", PathPrefix: "suggest/suggestion.suggestions", Reason: "engine suggestion variance"},
	}
	report := Report{
		Name: "suggest/suggestion",
		Mismatches: []Mismatch{
			{Path: "suggest/suggestion.suggestions", Want: []string{"a"}, Got: []string{"b"}},
			{Path: "suggest/suggestion.results[0].title", Want: "A", Got: "B"},
		},
	}
	filtered := gaps.Filter(report)
	require.Len(t, filtered.Mismatches, 1)
	require.Equal(t, "suggest/suggestion.results[0].title", filtered.Mismatches[0].Path)
	require.Len(t, filtered.Suppressed, 1)
}
