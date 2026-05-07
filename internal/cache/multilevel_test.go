package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/pkg/models"
)

func TestMultiLevelCache(t *testing.T) {
	// Use nil redis (memory-only) for testing
	c, err := NewMultiLevel("")
	require.NoError(t, err)

	resp := &models.Response{
		Query: "test",
		Results: []models.Result{
			{Title: "Test Result", URL: "https://example.com"},
		},
	}

	// Set and get
	c.Set("test-key", resp, time.Minute)
	got, ok := c.Get("test-key")
	require.True(t, ok)
	assert.Equal(t, resp.Query, got.Query)
	assert.Len(t, got.Results, 1)

	// Delete
	c.Delete("test-key")
	_, ok = c.Get("test-key")
	assert.False(t, ok)
}
