package wikipedia

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
)

func TestWikipediaEngine(t *testing.T) {
	w := &Wikipedia{}
	ok := w.Init(context.Background(), engine.EngineInitConfig{})
	assert.True(t, ok)
	assert.Equal(t, "wikipedia", w.Name())
}
