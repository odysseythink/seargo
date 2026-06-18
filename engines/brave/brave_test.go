package brave

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
)

func TestBraveEngine(t *testing.T) {
	b := &Brave{}
	err := b.Init(nil, engine.EngineInitConfig{})
	assert.NoError(t, err)
	assert.Equal(t, "brave", b.Name())
}
