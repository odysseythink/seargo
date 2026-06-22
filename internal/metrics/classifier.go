package metrics

import (
	"errors"
	"strings"

	seerrors "github.com/seargo/seargo/internal/errors"
)

// ClassifyError maps an engine search error to one of 5 metric classes.
// Returns "" if err is nil.
func ClassifyError(err error) string {
	if err == nil {
		return ""
	}

	// 1. Check typed errors (supports wrapping via errors.As)
	var ee *seerrors.EngineError
	if errors.As(err, &ee) {
		switch ee.SuspendedTimeCategory {
		case "captcha":
			return "captcha"
		case "access_denied":
			return "access-denied"
		}
	}

	var ae *seerrors.AppError
	if errors.As(err, &ae) {
		if ae.Code == "REQUEST_TIMEOUT" || ae.Code == "CONNECTION_FAILED" {
			return "timeout"
		}
		if ae.Code == "PROXY_ERROR" || ae.Code == "HTTP_ERROR" {
			return "http-error"
		}
	}

	// 2. Message-based classification
	msg := strings.ToLower(err.Error())

	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") {
		return "timeout"
	}

	if strings.Contains(msg, "captcha") {
		return "captcha"
	}

	if strings.Contains(msg, "403") || strings.Contains(msg, "access denied") {
		return "access-denied"
	}

	// "parse" as a standalone word — avoid matching words like "sparse"
	if strings.Contains(msg, "parse error") || strings.Contains(msg, "parsing") || strings.Contains(msg, "parser") {
		return "parse-error"
	}

	return "http-error"
}
