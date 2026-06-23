package upstream

import (
	"testing"
	"time"

	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestNormalize_BothShapes(t *testing.T) {
	up := &UpstreamResponse{
		Query: "golang",
		Results: []UpstreamResult{
			{
				URL:       "https://go.dev/doc",
				Title:     "Documentation - The Go Programming Language",
				Content:   "Go docs",
				Engine:    "google",
				Engines:   []string{"google"},
				Category:  "general",
				Template:  "default.html",
				Score:     1.0,
				Positions: []int{1},
			},
		},
		Suggestions: []string{"golang tutorial"},
	}

	published := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sg := &models.Response{
		Query: "golang",
		Results: []models.Result{
			{
				URL:         "https://go.dev/doc",
				Title:       "Documentation - The Go Programming Language",
				Content:     "Go docs",
				Engine:      "google",
				Engines:     []string{"google"},
				Category:    models.CategoryGeneral,
				Template:    "default.html",
				Score:       1.0,
				Positions:   []int{1},
				PublishedAt: &published,
			},
		},
		Suggestions: []string{"golang tutorial"},
	}

	nup := NormalizeUpstream(up)
	nsg := NormalizeSearGo(sg)

	require.Equal(t, nup.Query, nsg.Query)
	require.Len(t, nsg.Results, 1)
	require.Equal(t, "https://go.dev/doc", nsg.Results[0].URL)
	require.Equal(t, "2024-01-01", nsg.Results[0].PublishedDate)
	require.Equal(t, []string{"golang tutorial"}, nsg.Suggestions)
}

func TestNormalize_TypedFields(t *testing.T) {
	up := &UpstreamResponse{
		Results: []UpstreamResult{
			{
				URL:       "https://example.com/img.png",
				Template:  "images.html",
				ImgSrc:    "https://example.com/img.png",
				Thumbnail: "https://example.com/thumb.png",
			},
		},
	}
	n := NormalizeUpstream(up)
	require.Equal(t, "images.html", n.Results[0].Template)
	require.Equal(t, "https://example.com/img.png", n.Results[0].TypedFields["img_src"])
	require.Equal(t, "https://example.com/thumb.png", n.Results[0].TypedFields["thumbnail"])
}
