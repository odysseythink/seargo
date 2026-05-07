package duckduckgo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/pkg/models"
)

func TestDuckDuckGoSearch(t *testing.T) {
	ddg := &DuckDuckGo{}
	err := ddg.Init(nil)
	require.NoError(t, err)

	resp, err := ddg.Search(context.Background(), &models.Request{
		Query:    "go programming language",
		Category: models.CategoryGeneral,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Greater(t, len(resp.Results), 0, "Expected some results from DuckDuckGo")

	if len(resp.Results) > 0 {
		assert.NotEmpty(t, resp.Results[0].Title)
		assert.NotEmpty(t, resp.Results[0].URL)
	}
}
