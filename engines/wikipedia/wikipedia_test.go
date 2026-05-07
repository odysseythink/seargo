package wikipedia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWikipediaEngine(t *testing.T) {
	w := &Wikipedia{}
	err := w.Init(nil)
	assert.NoError(t, err)
	assert.Equal(t, "wikipedia", w.Name())
}
