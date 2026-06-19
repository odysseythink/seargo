package httpx

import (
	"strings"

	seerrors "github.com/seargo/seargo/internal/errors"
)

// raiseForHTTPError classifies HTTP responses with status >= 400 into
// known error patterns (Cloudflare JS challenge, Cloudflare CAPTCHA,
// Google reCAPTCHA, access denied, rate limiting, generic HTTP error).
func raiseForHTTPError(resp *Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	// Lowercase body for case-insensitive matching
	body := strings.ToLower(string(resp.Body))

	// Cloudflare JS challenge (503 + specific patterns)
	if resp.StatusCode == 503 || resp.StatusCode == 429 {
		if strings.Contains(body, "__cf_chl_jschl_tk__=") {
			return seerrors.EngineCaptchaError.WithMessage("Cloudflare JS challenge")
		}
		if strings.Contains(body, "/cdn-cgi/challenge-platform/") &&
			strings.Contains(body, "orchestrate/jsch/v1") &&
			strings.Contains(body, "window._cf_chl_enter(") {
			return seerrors.EngineCaptchaError.WithMessage("Cloudflare CAPTCHA challenge")
		}
	}

	// Cloudflare CAPTCHA at 403
	if resp.StatusCode == 403 && strings.Contains(body, "__cf_chl_captcha_tk__=") {
		return seerrors.EngineCaptchaError.WithMessage("Cloudflare CAPTCHA")
	}

	// Cloudflare Firewall 1020
	if resp.StatusCode == 403 && strings.Contains(body, "cf-error-code\">1020") {
		return seerrors.EngineAccessDeniedError.WithMessage("Cloudflare Firewall (1020)")
	}

	// Google reCAPTCHA
	if resp.StatusCode == 503 && strings.Contains(body, "https://www.google.com/recaptcha/") {
		return seerrors.EngineCaptchaError.WithMessage("Google reCAPTCHA")
	}

	// 402, 403 → Access Denied
	if resp.StatusCode == 402 || resp.StatusCode == 403 {
		return seerrors.EngineAccessDeniedError.WithMessage("HTTP " + statusText(resp.StatusCode))
	}

	// 429 → Too Many Requests
	if resp.StatusCode == 429 {
		return seerrors.EngineTooManyRequestsError.WithMessage("HTTP 429 Too Many Requests")
	}

	// Generic HTTP error
	return seerrors.HTTPError.WithMessage("HTTP " + statusText(resp.StatusCode))
}

func statusText(code int) string {
	switch code {
	case 400:
		return "400 Bad Request"
	case 401:
		return "401 Unauthorized"
	case 402:
		return "402 Payment Required"
	case 403:
		return "403 Forbidden"
	case 404:
		return "404 Not Found"
	case 405:
		return "405 Method Not Allowed"
	case 429:
		return "429 Too Many Requests"
	case 500:
		return "500 Internal Server Error"
	case 502:
		return "502 Bad Gateway"
	case 503:
		return "503 Service Unavailable"
	case 504:
		return "504 Gateway Timeout"
	default:
		return string(rune(code))
	}
}

// statusClass returns a string label for the HTTP status code range.
func statusClass(code int) string {
	if code == 0 {
		return "error"
	}
	if code >= 200 && code < 300 {
		return "2xx"
	}
	if code >= 300 && code < 400 {
		return "3xx"
	}
	if code >= 400 && code < 500 {
		return "4xx"
	}
	if code >= 500 {
		return "5xx"
	}
	return "other"
}

// errorClass returns a short label for error classification in metrics.
func errorClass(err error) string {
	if err == nil {
		return ""
	}
	switch err.(type) {
	case *seerrors.EngineError:
		ee := err.(*seerrors.EngineError)
		switch ee.SuspendedTimeCategory {
		case "captcha":
			return "captcha"
		case "access_denied":
			return "access_denied"
		case "too_many_requests":
			return "too_many_requests"
		}
		return "engine_error"
	}
	if ae, ok := err.(*seerrors.AppError); ok {
		switch ae.Code {
		case "REQUEST_TIMEOUT":
			return "timeout"
		case "CONNECTION_FAILED":
			return "connection"
		case "PROXY_ERROR":
			return "proxy"
		}
	}
	return "other"
}
