package httpx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	seerrors "github.com/seargo/seargo/internal/errors"
)

func TestComputeBackoff_Base(t *testing.T) {
	b := computeBackoff(0, 100*time.Millisecond, 2*time.Second)
	assert.GreaterOrEqual(t, b, time.Duration(0))
	assert.LessOrEqual(t, b, 100*time.Millisecond, "first attempt: max = base * 2^0 = 100ms")
}

func TestComputeBackoff_SecondAttempt(t *testing.T) {
	b := computeBackoff(1, 100*time.Millisecond, 2*time.Second)
	assert.GreaterOrEqual(t, b, time.Duration(0))
	assert.LessOrEqual(t, b, 200*time.Millisecond, "second attempt: max = base * 2^1 = 200ms")
}

func TestComputeBackoff_CappedAtMax(t *testing.T) {
	for i := 0; i < 20; i++ {
		b := computeBackoff(10, 500*time.Millisecond, 2*time.Second)
		assert.LessOrEqual(t, b, 2*time.Second, "should never exceed max delay")
	}
}

func TestComputeBackoff_JitterRange(t *testing.T) {
	seen := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		b := computeBackoff(5, 100*time.Millisecond, 2*time.Second)
		seen[b] = true
	}
	assert.Greater(t, len(seen), 1, "jitter should produce varied delays")
}

func TestShouldRetryHTTPError_Nil(t *testing.T) {
	assert.False(t, shouldRetryHTTPError(503, nil))
}

func TestShouldRetryHTTPError_False(t *testing.T) {
	assert.False(t, shouldRetryHTTPError(503, false))
}

func TestShouldRetryHTTPError_True(t *testing.T) {
	assert.True(t, shouldRetryHTTPError(503, true))
	assert.True(t, shouldRetryHTTPError(404, true))
	assert.False(t, shouldRetryHTTPError(200, true))
}

func TestShouldRetryHTTPError_Int(t *testing.T) {
	assert.True(t, shouldRetryHTTPError(503, 503))
	assert.False(t, shouldRetryHTTPError(502, 503))
}

func TestShouldRetryHTTPError_List(t *testing.T) {
	list := []interface{}{403, 429, 503}
	assert.True(t, shouldRetryHTTPError(503, list))
	assert.True(t, shouldRetryHTTPError(429, list))
	assert.False(t, shouldRetryHTTPError(502, list))
}

func TestNetwork_Request_RetryOnTransportError(t *testing.T) {
	err := seerrors.ConnectionFailedError.WithMessage("connection refused")
	assert.True(t, isRetryableTransportError(err))
}

func TestNetwork_Request_NotRetryable_EngineError(t *testing.T) {
	err := seerrors.EngineCaptchaError.WithMessage("captcha")
	assert.False(t, isRetryableTransportError(err))

	err2 := seerrors.EngineAccessDeniedError.WithMessage("denied")
	assert.False(t, isRetryableTransportError(err2))

	err3 := seerrors.EngineTooManyRequestsError.WithMessage("429")
	assert.False(t, isRetryableTransportError(err3))
}

func TestNetwork_Request_Retryable_Timeout(t *testing.T) {
	assert.True(t, isRetryableTransportError(seerrors.RequestTimeoutError))
	assert.True(t, isRetryableTransportError(seerrors.ProxyError))
	assert.True(t, isRetryableTransportError(seerrors.ConnectionFailedError))
}

func TestNetwork_Request_NotRetryable_ContextCanceled(t *testing.T) {
	assert.False(t, isRetryableTransportError(context.Canceled))
	assert.False(t, isRetryableTransportError(context.DeadlineExceeded))
}

func TestNetwork_IsRetryableFull(t *testing.T) {
	// Combined: transport error vs HTTP error vs engine error
	assert.True(t, isRetryable(seerrors.ConnectionFailedError, nil, 0, 1))

	// engine captcha → not retryable
	assert.False(t, isRetryable(seerrors.EngineCaptchaError, nil, 0, 1))

	// HTTP 503 with policy=true → retryable
	resp := &Response{StatusCode: 503}
	assert.True(t, isRetryable(errors.New("http"), resp, 0, 1))

	// Attempt >= retries → not retryable
	assert.False(t, isRetryable(seerrors.ConnectionFailedError, nil, 1, 1))
}
