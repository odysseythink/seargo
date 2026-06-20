package yahoo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
)

func TestYahooEngine(t *testing.T) {
	y := &Yahoo{}
	ok := y.Init(context.Background(), engine.EngineInitConfig{})
	assert.True(t, ok)
	assert.Equal(t, "yahoo", y.Name())
}
