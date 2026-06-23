package upstream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiff_DetectsMismatch(t *testing.T) {
	want := NormalizedResponse{
		Query: "golang",
		Results: []NormalizedResult{
			{URL: "https://go.dev/", Title: "Go"},
		},
		Suggestions: []string{"golang tutorial"},
	}
	got := NormalizedResponse{
		Query: "golang",
		Results: []NormalizedResult{
			{URL: "https://go.dev/", Title: "GoLang"},
		},
		Suggestions: []string{"golang tour"},
	}

	diffs := Diff("test", want, got)
	require.Len(t, diffs, 2)
	require.Equal(t, "test.results[0].title", diffs[0].Path)
	require.Equal(t, "Go", diffs[0].Want)
	require.Equal(t, "GoLang", diffs[0].Got)
	require.Equal(t, "test.suggestions", diffs[1].Path)
}

func TestDiff_TypedFields(t *testing.T) {
	want := NormalizedResponse{
		Results: []NormalizedResult{
			{URL: "https://example.com/", Template: "images.html", TypedFields: map[string]string{"img_src": "https://example.com/a.png"}},
		},
	}
	got := NormalizedResponse{
		Results: []NormalizedResult{
			{URL: "https://example.com/", Template: "images.html", TypedFields: map[string]string{"img_src": "https://example.com/b.png"}},
		},
	}
	ms := Diff("tf", want, got)
	require.Len(t, ms, 1)
	require.Equal(t, "tf.results[0].typed_fields.img_src", ms[0].Path)
}
