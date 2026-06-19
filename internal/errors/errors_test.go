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

func TestEngineError_WithDetails(t *testing.T) {
	e := EngineCaptchaError.WithDetails("test")
	assert.Contains(t, e.Error(), "ENGINE_CAPTCHA")
	assert.Equal(t, 503, e.Status)
	assert.NotEmpty(t, e.Details)
}

func TestEngineError_WithMessage(t *testing.T) {
	e := EngineAccessDeniedError.WithMessage("access denied: 403")
	assert.Contains(t, e.Message, "access denied")
	assert.Equal(t, "ENGINE_ACCESS_DENIED", e.Code)
}

func TestEngineError_SentinelImmutability(t *testing.T) {
	orig := EngineCaptchaError.Message
	_ = EngineCaptchaError.WithMessage("temp")
	assert.Equal(t, orig, EngineCaptchaError.Message, "sentinel should be immutable")
}

func TestHTTPError(t *testing.T) {
	e := HTTPError.WithMessage("404 not found")
	assert.Contains(t, e.Message, "404")
	assert.Equal(t, 503, e.Status)
}

func TestRequestTimeoutError(t *testing.T) {
	assert.Contains(t, RequestTimeoutError.Code, "REQUEST_TIMEOUT")
	assert.Equal(t, 504, RequestTimeoutError.Status)
}

func TestAppError_WithMessage(t *testing.T) {
	e := ErrRateLimited.WithMessage("custom message")
	assert.Equal(t, "custom message", e.Message)
	assert.Equal(t, 429, e.Status)
	assert.Equal(t, "RATE_LIMITED", e.Code)
}
