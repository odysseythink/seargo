package brave

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBraveEngine(t *testing.T) {
	b := &Brave{}
	err := b.Init(nil)
	assert.NoError(t, err)
	assert.Equal(t, "brave", b.Name())
}
