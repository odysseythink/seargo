package duckduckgo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/pkg/models"
)

func TestDuckDuckGoEngine(t *testing.T) {
	ddg := &DuckDuckGo{}
	err := ddg.Init(nil, engine.EngineInitConfig{})
	assert.NoError(t, err)
	assert.Equal(t, "duckduckgo", ddg.Name())
	assert.Contains(t, ddg.Categories(), models.CategoryGeneral)
}
