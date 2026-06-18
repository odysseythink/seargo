package wikipedia

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
)

func TestWikipediaEngine(t *testing.T) {
	w := &Wikipedia{}
	err := w.Init(nil, engine.EngineInitConfig{})
	assert.NoError(t, err)
	assert.Equal(t, "wikipedia", w.Name())
}
