package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	err := Init("debug", "stdout")
	require.NoError(t, err)
	assert.NotNil(t, Default())
}

func TestWithContext(t *testing.T) {
	Init("debug", "stdout")
	ctx := context.WithValue(context.Background(), "request_id", "abc123")
	l := WithContext(ctx)
	assert.NotNil(t, l)
}
