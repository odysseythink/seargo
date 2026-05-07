package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	err := ErrInvalidRequest.WithDetails("missing query")
	assert.Equal(t, "INVALID_REQUEST", err.Code)
	assert.Equal(t, 400, err.Status)
	assert.Equal(t, "missing query", err.Details)
	assert.Contains(t, err.Error(), "INVALID_REQUEST")
}
