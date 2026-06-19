package httpx

import (
	"context"
	"math/rand"
	"time"

	seerrors "github.com/seargo/seargo/internal/errors"
)

// RetryPolicy configures retry behavior for a Network.
type RetryPolicy struct {
	MaxRetries       int
	BaseDelay        time.Duration
	MaxDelay         time.Duration
	RetryOnHTTPError interface{}
}

// computeBackoff calculates an exponential backoff delay with full jitter.
func computeBackoff(attempt int, base, max time.Duration) time.Duration {
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	if max <= 0 {
		max = 2 * time.Second
	}

	exp := base
	for i := 0; i < attempt; i++ {
		exp *= 2
	}
	if exp > max {
		exp = max
	}

	if exp <= 0 {
		return 0
	}
	jitter := time.Duration(rand.Int63n(int64(exp)))
	return jitter
}

// shouldRetryHTTPError determines if an HTTP status code should trigger a retry.
func shouldRetryHTTPError(status int, spec interface{}) bool {
	if spec == nil {
		return false
	}
	switch v := spec.(type) {
	case bool:
		if v {
			return status >= 400 && status <= 599
		}
		return false
	case int:
		return status == v
	case float64:
		return status == int(v)
	case []interface{}:
		for _, item := range v {
			switch iv := item.(type) {
			case int:
				if status == iv {
					return true
				}
			case float64:
				if status == int(iv) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

// isRetryableTransportError checks whether a transport-level error is retryable.
func isRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Engine-level errors are not retryable
	if _, ok := err.(*seerrors.EngineError); ok {
		return false
	}

	// Check for specific sentinel codes
	if ae, ok := err.(*seerrors.AppError); ok {
		if ae.Code == "REQUEST_TIMEOUT" || ae.Code == "CONNECTION_FAILED" || ae.Code == "PROXY_ERROR" {
			return true
		}
	}

	return true // generic transport errors are retryable
}

// isRetryable determines if a request should be retried given the error,
// response, current attempt count, and max retries.
func isRetryable(err error, resp *Response, attempt, maxRetries int) bool {
	if attempt >= maxRetries {
		return false
	}
	if err == nil {
		return false
	}

	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	if _, ok := err.(*seerrors.EngineError); ok {
		return false
	}

	if isRetryableTransportError(err) {
		return true
	}

	// HTTP errors with a response
	if resp != nil && resp.StatusCode >= 400 {
		return true
	}

	return false
}
