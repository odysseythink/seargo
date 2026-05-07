package yahoo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYahooEngine(t *testing.T) {
	y := &Yahoo{}
	err := y.Init(nil)
	assert.NoError(t, err)
	assert.Equal(t, "yahoo", y.Name())
}
