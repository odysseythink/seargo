package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/pkg/models"
)

func TestTypedResultContainer_AddPluginResults(t *testing.T) {
	c := NewTypedResultContainer(map[string]float64{"google": 1.0})

	c.Extend("google", []models.Result{
		{Kind: "main", Title: "web result", URL: "https://example.com/a"},
	}, 0)

	c.AddPluginResults([]models.Result{
		{Kind: "answer", Title: "instant answer", Content: "42"},
		{Kind: "keyvalue", Title: "config table", URL: "https://example.com/kv"},
	})

	c.Close()

	// Answers go to GetAnswers
	answers := c.GetAnswers()
	assert.Len(t, answers, 1)
	assert.Equal(t, "42", answers[0].Answer)

	// Main and keyvalue go to Results
	results := c.Results()
	var foundMain, foundKeyValue bool
	for _, r := range results {
		if r.Kind == "main" && r.Title == "web result" {
			foundMain = true
		}
		if r.Kind == "keyvalue" && r.Title == "config table" {
			foundKeyValue = true
		}
	}
	assert.True(t, foundMain, "engine result should still be in output")
	assert.True(t, foundKeyValue, "plugin keyvalue result should be in output")
}

func TestTypedResultContainer_AddPluginResults_EmptyList(t *testing.T) {
	c := NewTypedResultContainer(map[string]float64{"google": 1.0})
	c.AddPluginResults(nil)
	c.AddPluginResults([]models.Result{})
	c.Close()
	assert.Empty(t, c.Results())
	assert.Empty(t, c.GetAnswers())
}

func TestTypedResultContainer_AddPluginResults_MultipleAnswers(t *testing.T) {
	c := NewTypedResultContainer(map[string]float64{})
	c.AddPluginResults([]models.Result{
		{Kind: "answer", Title: "answer A", Content: "text A", URL: "https://a.example.com"},
		{Kind: "answer", Title: "answer B", Content: "text B", URL: "https://b.example.com"},
	})
	c.Close()
	answers := c.GetAnswers()
	assert.Len(t, answers, 2)
}

func TestTypedResultContainer_AddPluginResults_WithSuggestions(t *testing.T) {
	c := NewTypedResultContainer(map[string]float64{})
	c.AddPluginResults([]models.Result{
		{Kind: "suggestion", Title: "did you mean X?"},
	})
	c.Close()
	suggestions := c.GetSuggestions()
	assert.Len(t, suggestions, 1)
}

func TestTypedResultContainer_AddPluginResults_WithInfobox(t *testing.T) {
	c := NewTypedResultContainer(map[string]float64{})
	c.AddPluginResults([]models.Result{
		{Kind: "infobox", Title: "Wikipedia", Content: "summary", URL: "https://en.wikipedia.org/wiki/X"},
	})
	c.Close()
	infoboxes := c.GetInfoboxes()
	assert.Len(t, infoboxes, 1)
}
