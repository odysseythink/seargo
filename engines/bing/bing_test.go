package bing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBingEngine(t *testing.T) {
	b := &Bing{}
	err := b.Init(nil)
	assert.NoError(t, err)
	assert.Equal(t, "bing", b.Name())
}
