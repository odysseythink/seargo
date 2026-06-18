package bing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
)

func TestBingEngine(t *testing.T) {
	b := &Bing{}
	err := b.Init(nil, engine.EngineInitConfig{})
	assert.NoError(t, err)
	assert.Equal(t, "bing", b.Name())
}
