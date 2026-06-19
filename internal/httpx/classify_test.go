package httpx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	seerrors "github.com/seargo/seargo/internal/errors"
)

func TestRaiseForHTTPError_Success(t *testing.T) {
	resp := &Response{StatusCode: 200, Body: []byte("ok")}
	assert.NoError(t, raiseForHTTPError(resp))
}

func TestRaiseForHTTPError_CloudflareChallenge_503(t *testing.T) {
	body := `<html><head><script>/cdn-cgi/challenge-platform/orchestrate/jsch/v1</script>` +
		`<script>window._cf_chl_enter(</script></head></html>`
	resp := &Response{StatusCode: 503, Body: []byte(body)}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cloudflare")
	assert.Contains(t, err.Error(), "ENGINE_CAPTCHA")
}

func TestRaiseForHTTPError_CloudflareCaptcha_403(t *testing.T) {
	body := `<html>__cf_chl_captcha_tk__=abc123</html>`
	resp := &Response{StatusCode: 403, Body: []byte(body)}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ENGINE_CAPTCHA")
}

func TestRaiseForHTTPError_Cloudflare1020(t *testing.T) {
	body := `<html><span class="cf-error-code">1020</span></html>`
	resp := &Response{StatusCode: 403, Body: []byte(body)}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ENGINE_ACCESS_DENIED")
}

func TestRaiseForHTTPError_Recaptcha(t *testing.T) {
	body := `<script src="https://www.google.com/recaptcha/api.js"></script>`
	resp := &Response{StatusCode: 503, Body: []byte(body)}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ENGINE_CAPTCHA")
}

func TestRaiseForHTTPError_429_TooManyRequests(t *testing.T) {
	resp := &Response{StatusCode: 429, Body: []byte("rate limited")}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ENGINE_TOO_MANY_REQUESTS")
}

func TestRaiseForHTTPError_403_AccessDenied(t *testing.T) {
	resp := &Response{StatusCode: 403, Body: []byte("forbidden")}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ENGINE_ACCESS_DENIED")
}

func TestRaiseForHTTPError_402_AccessDenied(t *testing.T) {
	resp := &Response{StatusCode: 402, Body: []byte("payment required")}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ENGINE_ACCESS_DENIED")
}

func TestRaiseForHTTPError_500_GenericHTTPError(t *testing.T) {
	resp := &Response{StatusCode: 500, Body: []byte("internal server error")}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP_ERROR")
}

func TestRaiseForHTTPError_Normal503_NotCaptcha(t *testing.T) {
	resp := &Response{StatusCode: 503, Body: []byte("<html><body>Service Unavailable</body></html>")}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP_ERROR", "normal 503 should not be classified as captcha")
	assert.NotContains(t, err.Error(), "ENGINE_CAPTCHA")
}

func TestRaiseForHTTPError_CaseInsensitive(t *testing.T) {
	body := `<html>__CF_CHL_CAPTCHA_TK__=abc</html>`
	resp := &Response{StatusCode: 403, Body: []byte(body)}
	err := raiseForHTTPError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ENGINE_CAPTCHA")
}

func TestStatusClass(t *testing.T) {
	assert.Equal(t, "2xx", statusClass(200))
	assert.Equal(t, "3xx", statusClass(301))
	assert.Equal(t, "4xx", statusClass(404))
	assert.Equal(t, "5xx", statusClass(503))
	assert.Equal(t, "error", statusClass(0))
}

func TestErrorClass(t *testing.T) {
	assert.Equal(t, "captcha", errorClass(seerrors.EngineCaptchaError))
	assert.Equal(t, "access_denied", errorClass(seerrors.EngineAccessDeniedError))
	assert.Equal(t, "too_many_requests", errorClass(seerrors.EngineTooManyRequestsError))
	assert.Equal(t, "timeout", errorClass(seerrors.RequestTimeoutError))
	assert.Equal(t, "connection", errorClass(seerrors.ConnectionFailedError))
	assert.Equal(t, "proxy", errorClass(seerrors.ProxyError))
	assert.Equal(t, "other", errorClass(seerrors.HTTPError))
	assert.Equal(t, "", errorClass(nil))
}

func TestClassifyTransportError_Timeout(t *testing.T) {
	err := classifyTransportError(seerrors.RequestTimeoutError)
	assert.Contains(t, err.Error(), "REQUEST_TIMEOUT")
}

func TestClassifyTransportError_ConnectionRefused(t *testing.T) {
	err := classifyTransportError(seerrors.ConnectionFailedError)
	assert.Contains(t, err.Error(), "CONNECTION_FAILED")
}

func TestClassifyTransportError_ProxyError(t *testing.T) {
	err := classifyTransportError(seerrors.ProxyError)
	assert.Contains(t, err.Error(), "PROXY_ERROR")
}

func TestClassifyTransportError_Generic(t *testing.T) {
	unknownErr := fmt.Errorf("unknown network glitch")
	err := classifyTransportError(unknownErr)
	assert.NotNil(t, err)
}

func TestRedactProxyURL(t *testing.T) {
	redacted := redactProxyURL("http://user:password@proxy.example.com:8080")
	assert.NotContains(t, redacted, "user")
	assert.NotContains(t, redacted, "password")
	assert.Contains(t, redacted, "proxy.example.com")

	clean := redactProxyURL("http://proxy.example.com:8080")
	assert.Equal(t, "http://proxy.example.com:8080", clean)

	assert.Equal(t, "", redactProxyURL(""))

	socks := redactProxyURL("socks5://admin:secret@tor:9050")
	assert.NotContains(t, socks, "admin")
	assert.NotContains(t, socks, "secret")
	assert.Contains(t, socks, "tor:9050")
}
