package duckduckgo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/pkg/models"
)

func TestDuckDuckGoEngine(t *testing.T) {
	ddg := &DuckDuckGo{}
	ok := ddg.Init(context.Background(), engine.EngineInitConfig{})
	assert.True(t, ok)
	assert.Equal(t, "duckduckgo", ddg.Name())
	assert.Contains(t, ddg.Categories(), models.CategoryGeneral)
}
