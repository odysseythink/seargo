package search

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/pkg/models"
)

func TestNormalizeURL_SchemeAndTrailingSlash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http://example.com/", "http://example.com"},
		{"https://example.com", "https://example.com"},
		{"HTTP://EXAMPLE.COM/", "http://example.com"},
		{"http://www.example.com/", "http://example.com"},
		{"https://www.example.com/path/", "https://example.com/path"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeURL(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeURL_TrackingParams(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://example.com/a?utm_source=x", "https://example.com/a"},
		{"https://example.com/a?utm_medium=email&b=2", "https://example.com/a?b=2"},
		{"https://example.com/a?fbclid=123", "https://example.com/a"},
		{"https://example.com/a?gclid=abc", "https://example.com/a"},
		{"https://example.com/a?ref=site", "https://example.com/a"},
		{"https://example.com/a?q=test", "https://example.com/a?q=test"},
		{"https://example.com/a?search=golang", "https://example.com/a?search=golang"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeURL(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeURL_InvalidURL(t *testing.T) {
	// Go's url.Parse is lenient; invalid URLs get partial normalization
	_ = normalizeURL("not a url")
}

func TestResultContainer_Extend(t *testing.T) {
	weights := map[string]float64{"google": 1.0, "bing": 2.0}
	rc := NewResultContainer(weights)

	rc.Extend("google", []models.Result{
		{Title: "Result 1", URL: "https://example.com/a", Content: "short"},
		{Title: "Result 2", URL: "https://example.com/b", Content: "text"},
	}, 0)

	rc.Close()
	results := rc.GetOrderedResults()
	assert.Len(t, results, 2)
	assert.Equal(t, []string{"google"}, results[0].Engines)
}

func TestResultContainer_Dedup(t *testing.T) {
	weights := map[string]float64{"google": 1.0, "bing": 1.0}
	rc := NewResultContainer(weights)

	rc.Extend("google", []models.Result{
		{Title: "Same", URL: "https://example.com/", Content: "from google"},
	}, 0)
	rc.Extend("bing", []models.Result{
		{Title: "Same but longer title", URL: "https://example.com", Content: "from bing longer"},
	}, 0)

	rc.Close()
	results := rc.GetOrderedResults()
	assert.Len(t, results, 1, "same URL should be deduped")
	assert.Equal(t, []string{"google", "bing"}, results[0].Engines)
	assert.Equal(t, "from bing longer", results[0].Content, "richest content wins")
	assert.Equal(t, "Same but longer title", results[0].Title, "richest title wins")
}

func TestResultContainer_DedupPreferHTTPS(t *testing.T) {
	weights := map[string]float64{"a": 1.0, "b": 1.0}
	rc := NewResultContainer(weights)

	rc.Extend("a", []models.Result{
		{Title: "X", URL: "http://example.com/path"},
	}, 0)
	rc.Extend("b", []models.Result{
		{Title: "X", URL: "https://example.com/path"},
	}, 0)

	rc.Close()
	results := rc.GetOrderedResults()
	assert.Len(t, results, 1)
	assert.Equal(t, "https://example.com/path", results[0].URL)
}

func TestResultContainer_NoDedupDifferentPaths(t *testing.T) {
	weights := map[string]float64{"a": 1.0}
	rc := NewResultContainer(weights)

	rc.Extend("a", []models.Result{
		{Title: "A", URL: "https://example.com/a"},
		{Title: "B", URL: "https://example.com/b"},
	}, 0)

	rc.Close()
	results := rc.GetOrderedResults()
	assert.Len(t, results, 2, "different paths should not merge")
}

func TestScoreCalculation(t *testing.T) {
	weights := map[string]float64{"google": 1.0, "bing": 2.0}
	rc := NewResultContainer(weights)

	rc.Extend("google", []models.Result{
		{Title: "R1", URL: "https://x.com/1"},
	}, 0)
	rc.Extend("bing", []models.Result{
		{Title: "R1", URL: "https://x.com/1"},
	}, 0)

	rc.Close()
	results := rc.GetOrderedResults()
	assert.Len(t, results, 1)

	// google weight=1 pos=1, bing weight=2 pos=1
	// score = (1/1 + 2/1) * 2 = 6.0
	assert.InDelta(t, 6.0, results[0].Score, 0.01)
}

func TestCategoryGrouping(t *testing.T) {
	weights := map[string]float64{"e1": 1.0}
	rc := NewResultContainer(weights)

	for i := 0; i < 5; i++ {
		rc.Extend("e1", []models.Result{
			{Title: fmt.Sprintf("G%d", i), URL: fmt.Sprintf("https://x.com/g%d", i), Category: models.CategoryGeneral},
		}, i*2)
	}
	for i := 0; i < 5; i++ {
		rc.Extend("e1", []models.Result{
			{Title: fmt.Sprintf("I%d", i), URL: fmt.Sprintf("https://x.com/i%d", i), Category: models.CategoryImages},
		}, i*2)
	}

	rc.Close()
	results := rc.GetOrderedResults()
	assert.Len(t, results, 10)

	foundGeneral := false
	foundImages := false
	for _, r := range results {
		if r.Category == models.CategoryGeneral {
			foundGeneral = true
		}
		if foundGeneral && r.Category == models.CategoryImages {
			foundImages = true
		}
	}
	assert.True(t, foundImages, "grouping should cluster same-category results")
}

func TestResultContainer_Suggestions(t *testing.T) {
	rc := NewResultContainer(nil)
	rc.AddSuggestions("google", []string{"s1", "S1", "s2"})
	rc.AddSuggestions("bing", []string{"s3", "s2"})

	suggs := rc.GetSuggestions()
	assert.Len(t, suggs, 3, "case-insensitive dedup")
	assert.Equal(t, []string{"s1", "s2", "s3"}, suggs)
}

func TestResultContainer_Answers(t *testing.T) {
	rc := NewResultContainer(nil)
	rc.AddAnswers("google", []models.Answer{{Answer: "42", URL: "https://x.com"}})

	answers := rc.GetAnswers()
	assert.Len(t, answers, 1)
	assert.Equal(t, "42", answers[0].Answer)
}

func TestResultContainer_Infoboxes(t *testing.T) {
	rc := NewResultContainer(nil)
	rc.AddInfoboxes("wiki", []models.Infobox{{Title: "Go", Content: "Programming language", Engine: "wiki"}})

	infos := rc.GetInfoboxes()
	assert.Len(t, infos, 1)
}

func TestResultContainer_EngineData(t *testing.T) {
	rc := NewResultContainer(nil)
	rc.AddEngineData("google", map[string]any{"results": 10})

	data := rc.GetEngineData()
	assert.Contains(t, data, "google.results")
}

func TestResultContainer_Unresponsive(t *testing.T) {
	rc := NewResultContainer(nil)
	rc.MarkUnresponsive("google", "SearxEngineAccessDenied")
	rc.MarkUnresponsive("bing", "timeout")

	unresp := rc.GetUnresponsive()
	assert.Len(t, unresp, 2)
	assert.Equal(t, "google", unresp[0].Name)
	assert.Equal(t, "SearxEngineAccessDenied", unresp[0].Reason)
}
