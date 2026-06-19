package errors

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
		Status:  e.Status,
	}
}

func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: msg,
		Details: e.Details,
		Status:  e.Status,
	}
}

var (
	ErrInternal         = &AppError{Code: "INTERNAL_ERROR", Message: "internal server error", Status: 500}
	ErrInvalidRequest   = &AppError{Code: "INVALID_REQUEST", Message: "invalid request", Status: 400}
	ErrInvalidCategory  = &AppError{Code: "INVALID_CATEGORY", Message: "invalid category", Status: 400}
	ErrAllEnginesFailed = &AppError{Code: "ALL_ENGINES_FAILED", Message: "all search engines failed", Status: 503}
	ErrRateLimited      = &AppError{Code: "RATE_LIMITED", Message: "too many requests", Status: 429}
	ErrNotFound         = &AppError{Code: "NOT_FOUND", Message: "resource not found", Status: 404}
)

// EngineError is an error caused by an upstream search engine, carrying
// a suspended-time category hint for the suspension tracker.
type EngineError struct {
	*AppError
	SuspendedTimeCategory string // which SuspendedTimesConfig field to use
}

var (
	EngineCaptchaError         = &EngineError{AppError: &AppError{Code: "ENGINE_CAPTCHA", Message: "search engine returned a CAPTCHA", Status: 503}, SuspendedTimeCategory: "captcha"}
	EngineAccessDeniedError    = &EngineError{AppError: &AppError{Code: "ENGINE_ACCESS_DENIED", Message: "search engine access denied", Status: 503}, SuspendedTimeCategory: "access_denied"}
	EngineTooManyRequestsError = &EngineError{AppError: &AppError{Code: "ENGINE_TOO_MANY_REQUESTS", Message: "search engine rate limited", Status: 503}, SuspendedTimeCategory: "too_many_requests"}
	HTTPError                  = &AppError{Code: "HTTP_ERROR", Message: "HTTP error", Status: 503}
	RequestTimeoutError        = &AppError{Code: "REQUEST_TIMEOUT", Message: "request timeout", Status: 504}
	ConnectionFailedError      = &AppError{Code: "CONNECTION_FAILED", Message: "connection failed", Status: 503}
	ProxyError                 = &AppError{Code: "PROXY_ERROR", Message: "proxy error", Status: 503}
)

// WithMessage returns a new EngineError with the message replaced.
// The original sentinel is never mutated.
func (e *EngineError) WithMessage(msg string) *EngineError {
	app := *e.AppError
	app.Message = msg
	return &EngineError{AppError: &app, SuspendedTimeCategory: e.SuspendedTimeCategory}
}

// WithDetails returns a new EngineError with details set.
func (e *EngineError) WithDetails(details any) *EngineError {
	app := *e.AppError
	app.Details = details
	return &EngineError{AppError: &app, SuspendedTimeCategory: e.SuspendedTimeCategory}
}
