package metrics

import (
	"errors"
	"testing"

	seerrors "github.com/seargo/seargo/internal/errors"
)

func TestClassifyErrorTimeout(t *testing.T) {
	err := seerrors.RequestTimeoutError.WithMessage("request timeout")
	class := ClassifyError(err)
	if class != "timeout" {
		t.Errorf("expected 'timeout', got %q", class)
	}
	class = ClassifyError(errors.New("context deadline exceeded"))
	if class != "timeout" {
		t.Errorf("expected 'timeout' for deadline exceeded, got %q", class)
	}
}

func TestClassifyErrorCaptcha(t *testing.T) {
	err := seerrors.EngineCaptchaError.WithMessage("CAPTCHA detected")
	class := ClassifyError(err)
	if class != "captcha" {
		t.Errorf("expected 'captcha', got %q", class)
	}
	class = ClassifyError(errors.New("please solve captcha to continue"))
	if class != "captcha" {
		t.Errorf("expected 'captcha' for message match, got %q", class)
	}
	class = ClassifyError(errors.New("engine captain"))
	if class == "captcha" {
		t.Errorf("'captain' should not match 'captcha'")
	}
}

func TestClassifyErrorAccessDenied(t *testing.T) {
	err := seerrors.EngineAccessDeniedError.WithMessage("access denied")
	class := ClassifyError(err)
	if class != "access-denied" {
		t.Errorf("expected 'access-denied', got %q", class)
	}
	class = ClassifyError(errors.New("HTTP 403 Forbidden"))
	if class != "access-denied" {
		t.Errorf("expected 'access-denied' for 403, got %q", class)
	}
}

func TestClassifyErrorParse(t *testing.T) {
	class := ClassifyError(errors.New("parse error: unexpected token"))
	if class != "parse-error" {
		t.Errorf("expected 'parse-error', got %q", class)
	}
	class = ClassifyError(errors.New("sparse matrix overflow"))
	if class == "parse-error" {
		t.Errorf("'sparse' should not match 'parse'")
	}
}

func TestClassifyErrorHTTP(t *testing.T) {
	class := ClassifyError(errors.New("unknown network error"))
	if class != "http-error" {
		t.Errorf("expected 'http-error' for unknown, got %q", class)
	}
	class = ClassifyError(errors.New("HTTP 500 Internal Server Error"))
	if class != "http-error" {
		t.Errorf("expected 'http-error' for 500, got %q", class)
	}
}

func TestClassifyErrorNil(t *testing.T) {
	class := ClassifyError(nil)
	if class != "" {
		t.Errorf("expected empty string for nil, got %q", class)
	}
}
