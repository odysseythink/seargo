package google

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
)

func TestGoogleEngine(t *testing.T) {
	g := &Google{}
	err := g.Init(nil, engine.EngineInitConfig{})
	assert.NoError(t, err)
	assert.Equal(t, "google", g.Name())
}
