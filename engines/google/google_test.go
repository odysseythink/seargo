package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoogleEngine(t *testing.T) {
	g := &Google{}
	err := g.Init(nil)
	assert.NoError(t, err)
	assert.Equal(t, "google", g.Name())
}
