package httpx

import (
	"testing"

	"github.com/stretchr/testify/assert"

	seerrors "github.com/seargo/seargo/internal/errors"
)

func TestRecordMetrics(t *testing.T) {
	assert.NotPanics(t, func() {
		recordMetrics("default", "google", 200, 0, nil)
	})
	assert.NotPanics(t, func() {
		recordMetrics("default", "google", 503, 0, seerrors.EngineCaptchaError)
	})
}

func TestLogResponse_Debug_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		logResponse("google", "default", "GET", "https://example.com/search?q=test", 200, nil)
	})
}

func TestLogResponse_Info_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		logResponse("bing", "default", "POST", "https://example.com/api", 403, seerrors.EngineAccessDeniedError)
	})
}

func TestLogResponse_InfoOnlyHost(t *testing.T) {
	host := parseHost("https://example.com/search?q=secret")
	assert.Equal(t, "example.com", host)

	host2 := parseHost("http://sub.domain.com:8080/path?query=1")
	assert.Equal(t, "sub.domain.com", host2)
}

func TestResponseSizeLimit(t *testing.T) {
	assert.Greater(t, maxResponseSize, 0)
}

func TestMaxRequestSize(t *testing.T) {
	assert.Greater(t, maxRequestSize, 0)
}
