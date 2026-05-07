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

var (
	ErrInternal         = &AppError{Code: "INTERNAL_ERROR", Message: "internal server error", Status: 500}
	ErrInvalidRequest   = &AppError{Code: "INVALID_REQUEST", Message: "invalid request", Status: 400}
	ErrInvalidCategory  = &AppError{Code: "INVALID_CATEGORY", Message: "invalid category", Status: 400}
	ErrAllEnginesFailed = &AppError{Code: "ALL_ENGINES_FAILED", Message: "all search engines failed", Status: 503}
	ErrRateLimited      = &AppError{Code: "RATE_LIMITED", Message: "too many requests", Status: 429}
	ErrNotFound         = &AppError{Code: "NOT_FOUND", Message: "resource not found", Status: 404}
)
