package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/config"
)

func TestSuspensionTrackerBan(t *testing.T) {
	cfg := config.SearchConfig{
		BanTimeOnFail:    0.1,
		MaxBanTimeOnFail: 1.0,
		SuspendedTimes: config.SuspendedTimesConfig{
			SearxEngineAccessDenied:    3600,
			SearxEngineCaptcha:         7200,
			SearxEngineTooManyRequests: 60,
		},
	}

	st := NewSuspensionTracker(cfg)
	assert.False(t, st.IsSuspended("google"))

	st.Ban("google", "SearxEngineAccessDenied")
	assert.True(t, st.IsSuspended("google"))

	assert.False(t, st.IsSuspended("bing"))
}

func TestSuspensionTrackerBanExpiry(t *testing.T) {
	cfg := config.SearchConfig{
		BanTimeOnFail:    0.05,
		MaxBanTimeOnFail: 0.5,
	}

	st := NewSuspensionTracker(cfg)
	st.Ban("google", "SearxEngineTooManyRequests")
	assert.True(t, st.IsSuspended("google"))

	time.Sleep(100 * time.Millisecond)
	assert.False(t, st.IsSuspended("google"), "ban should expire after small duration")
}

func TestSuspensionTrackerClear(t *testing.T) {
	cfg := config.SearchConfig{
		BanTimeOnFail:    5.0,
		MaxBanTimeOnFail: 120.0,
	}

	st := NewSuspensionTracker(cfg)
	st.Ban("google", "SearxEngineCaptcha")
	assert.True(t, st.IsSuspended("google"))

	st.Clear("google")
	assert.False(t, st.IsSuspended("google"))
}

func TestSuspensionEscalatingBan(t *testing.T) {
	cfg := config.SearchConfig{
		BanTimeOnFail:    0.01,
		MaxBanTimeOnFail: 0.1,
	}

	st := NewSuspensionTracker(cfg)
	st.Ban("google", "SearxEngineTooManyRequests")
	assert.True(t, st.IsSuspended("google"))

	time.Sleep(20 * time.Millisecond)
	assert.False(t, st.IsSuspended("google"))
	st.Ban("google", "SearxEngineTooManyRequests")
	assert.True(t, st.IsSuspended("google"))

	time.Sleep(40 * time.Millisecond)
	assert.False(t, st.IsSuspended("google"))
	st.Ban("google", "SearxEngineTooManyRequests")
	assert.True(t, st.IsSuspended("google"))
}

func TestClassifyError(t *testing.T) {
	assert.Equal(t, "SearxEngineAccessDenied", classifyError(accessDeniedError()))
	assert.Equal(t, "SearxEngineCaptcha", classifyError(captchaError()))
	assert.Equal(t, "SearxEngineTooManyRequests", classifyError(tooManyError()))
	assert.Equal(t, "SearxEngineTooManyRequests", classifyError(unknownError()))
}

func accessDeniedError() error  { return &testError{msg: "access denied"} }
func captchaError() error       { return &testError{msg: "captcha required"} }
func tooManyError() error       { return &testError{msg: "too many requests"} }
func unknownError() error       { return &testError{msg: "unknown"} }

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
