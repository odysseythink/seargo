package search

import (
	"strings"
	"sync"
	"time"

	"github.com/seargo/seargo/internal/config"
)

type SuspensionTracker struct {
	mu     sync.RWMutex
	bans   map[string]banEntry
	config config.SearchConfig
}

type banEntry struct {
	reason string
	until  time.Time
	count  int
}

func NewSuspensionTracker(cfg config.SearchConfig) *SuspensionTracker {
	return &SuspensionTracker{
		bans:   make(map[string]banEntry),
		config: cfg,
	}
}

func (st *SuspensionTracker) Ban(engineName, errorClass string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	entry := st.bans[engineName]
	entry.count++
	entry.reason = errorClass

	// Check if there's a specific duration for this error class
	var duration float64
	if specificDuration := st.getSuspensionDuration(errorClass); specificDuration > 0 {
		duration = specificDuration
	} else {
		// Escalating ban: base_time * count, capped at max_ban_time
		duration = st.config.BanTimeOnFail * float64(entry.count)
		if duration > st.config.MaxBanTimeOnFail {
			duration = st.config.MaxBanTimeOnFail
		}
	}
	entry.until = time.Now().Add(time.Duration(duration * float64(time.Second)))

	st.bans[engineName] = entry
}

func (st *SuspensionTracker) getSuspensionDuration(errorClass string) float64 {
	switch errorClass {
	case "SearxEngineAccessDenied":
		return st.config.SuspendedTimes.SearxEngineAccessDenied
	case "SearxEngineCaptcha":
		return st.config.SuspendedTimes.SearxEngineCaptcha
	case "SearxEngineTooManyRequests":
		return st.config.SuspendedTimes.SearxEngineTooManyRequests
	case "cf_SearxEngineCaptcha":
		return st.config.SuspendedTimes.CfSearxEngineCaptcha
	case "cf_SearxEngineAccessDenied":
		return st.config.SuspendedTimes.CfSearxEngineAccessDenied
	case "recaptcha_SearxEngineCaptcha":
		return st.config.SuspendedTimes.RecaptchaSearxEngineCaptcha
	}
	return 0
}

func (st *SuspensionTracker) IsSuspended(engineName string) bool {
	st.mu.RLock()
	defer st.mu.RUnlock()

	entry, ok := st.bans[engineName]
	if !ok {
		return false
	}
	if time.Now().After(entry.until) {
		return false
	}
	return true
}

func (st *SuspensionTracker) Clear(engineName string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.bans, engineName)
}

func classifyError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())

	if strings.Contains(msg, "403") || strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "access denied") {
		return "SearxEngineAccessDenied"
	}
	if strings.Contains(msg, "captcha") || strings.Contains(msg, "recaptcha") ||
		strings.Contains(msg, "challenge") {
		return "SearxEngineCaptcha"
	}
	if strings.Contains(msg, "429") || strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "rate limit") {
		return "SearxEngineTooManyRequests"
	}

	return "SearxEngineTooManyRequests"
}
