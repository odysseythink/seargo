package engine

import "fmt"

// SearxEngineResponseException is a structured error for upstream engine
// HTTP responses that indicate a problem. Ported from SearXNG's
// SearxEngineResponseException.
type SearxEngineResponseException struct {
	EngineName string
	Message    string
	StatusCode int
	errorClass string
}

func (e *SearxEngineResponseException) Error() string {
	return fmt.Sprintf("[%s] %s (HTTP %d)", e.EngineName, e.Message, e.StatusCode)
}

// ErrorClass returns the suspension category for this error.
func (e *SearxEngineResponseException) ErrorClass() string {
	return e.errorClass
}

// NewSearxEngineResponseException creates a generic engine response error.
func NewSearxEngineResponseException(engineName, message string, statusCode int) *SearxEngineResponseException {
	return &SearxEngineResponseException{
		EngineName: engineName,
		Message:    message,
		StatusCode: statusCode,
		errorClass: classifyStatus(statusCode),
	}
}

// NewEngineAccessDeniedError creates an access-denied error.
func NewEngineAccessDeniedError(engineName string, statusCode int) *SearxEngineResponseException {
	return &SearxEngineResponseException{
		EngineName: engineName,
		Message:    fmt.Sprintf("access denied for engine %s", engineName),
		StatusCode: statusCode,
		errorClass: "access_denied",
	}
}

// NewEngineCaptchaError creates a captcha error.
func NewEngineCaptchaError(engineName string, statusCode int) *SearxEngineResponseException {
	return &SearxEngineResponseException{
		EngineName: engineName,
		Message:    fmt.Sprintf("captcha required for engine %s", engineName),
		StatusCode: statusCode,
		errorClass: "captcha",
	}
}

// NewEngineTooManyRequestsError creates a rate-limit error.
func NewEngineTooManyRequestsError(engineName string, statusCode int) *SearxEngineResponseException {
	return &SearxEngineResponseException{
		EngineName: engineName,
		Message:    fmt.Sprintf("too many requests for engine %s", engineName),
		StatusCode: statusCode,
		errorClass: "too_many_requests",
	}
}

// NewEngineTimeoutError creates a timeout error.
func NewEngineTimeoutError(engineName string) *SearxEngineResponseException {
	return &SearxEngineResponseException{
		EngineName: engineName,
		Message:    fmt.Sprintf("timeout for engine %s", engineName),
		StatusCode: 0,
		errorClass: "timeout",
	}
}

// classifyStatus maps HTTP status codes to error classes.
func classifyStatus(code int) string {
	switch {
	case code == 429:
		return "too_many_requests"
	case code == 403:
		return "access_denied"
	case code == 503:
		return "captcha"
	default:
		return "error"
	}
}

// IsNoResultStatus returns true if the given HTTP status code is configured
// as a "no result" status for this engine.
func (cfg EngineInitConfig) IsNoResultStatus(code int) bool {
	for _, s := range cfg.NoResultForHTTPStatus {
		if s == code {
			return true
		}
	}
	return false
}
