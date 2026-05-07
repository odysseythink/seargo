package httpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	c := New("SearGo/1.0", 10*time.Second)
	assert.NotNil(t, c)
	assert.NotNil(t, c.R())
}
