package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearxEngineResponseException(t *testing.T) {
	err := NewSearxEngineResponseException("google", "access denied", 403)
	assert.Contains(t, err.Error(), "google")
	assert.Contains(t, err.Error(), "access denied")
	assert.Contains(t, err.Error(), "403")
	assert.Equal(t, "access_denied", err.ErrorClass())
}

func TestSearxEngineAccessDenied(t *testing.T) {
	err := NewEngineAccessDeniedError("bing", 403)
	assert.Equal(t, "access_denied", err.ErrorClass())
	assert.Equal(t, 403, err.StatusCode)
}

func TestSearxEngineCaptcha(t *testing.T) {
	err := NewEngineCaptchaError("google", 503)
	assert.Equal(t, "captcha", err.ErrorClass())
}

func TestSearxEngineTooManyRequests(t *testing.T) {
	err := NewEngineTooManyRequestsError("ddg", 429)
	assert.Equal(t, "too_many_requests", err.ErrorClass())
}

func TestSearxEngineTimeout(t *testing.T) {
	err := NewEngineTimeoutError("slow_engine")
	assert.Equal(t, "timeout", err.ErrorClass())
}

func TestSearxEngineResponseException_NoResultForHTTPStatus(t *testing.T) {
	cfg := EngineInitConfig{
		NoResultForHTTPStatus: []int{404},
	}
	assert.True(t, cfg.IsNoResultStatus(404))
	assert.False(t, cfg.IsNoResultStatus(500))
}

func TestEngineInitConfig_IsNoResultStatus_Nil(t *testing.T) {
	cfg := EngineInitConfig{}
	assert.False(t, cfg.IsNoResultStatus(404))
}
