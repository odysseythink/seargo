package brave

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
)

func TestBraveEngine(t *testing.T) {
	b := &Brave{}
	ok := b.Init(context.Background(), engine.EngineInitConfig{})
	assert.True(t, ok)
	assert.Equal(t, "brave", b.Name())
}
