package yahoo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
)

func TestYahooEngine(t *testing.T) {
	y := &Yahoo{}
	err := y.Init(nil, engine.EngineInitConfig{})
	assert.NoError(t, err)
	assert.Equal(t, "yahoo", y.Name())
}
