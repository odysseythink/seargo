package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/pkg/models"
)

func TestTypedResultContainer_TimeoutEngineNotInEnginesUsed(t *testing.T) {
	c := NewTypedResultContainer(map[string]float64{"google": 1.0, "duckduckgo": 1.0})

	c.Extend("google", []models.Result{{
		Title:    "Result",
		URL:      "https://example.com",
		Content:  "content",
		Engine:   "google",
		Category: models.CategoryGeneral,
	}}, 0)

	c.MarkUnresponsive("duckduckgo", "timeout")

	used := c.GetEnginesUsed()
	assert.Contains(t, used, "google")
	assert.NotContains(t, used, "duckduckgo", "timed-out engine must not appear in EnginesUsed")

	failed := c.GetEnginesFailed()
	assert.Contains(t, failed, "duckduckgo:timeout")
}
