package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/security"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestTrustedProxy_ClientIPInContext(t *testing.T) {
	trusted, _ := security.ParseProxyList([]string{"10.0.0.0/8"})
	ext := security.NewIPExtractor(trusted)

	r := gin.New()
	r.Use(TrustedProxy(ext))
	r.GET("/test", func(c *gin.Context) {
		ip, _ := c.Get("clientIP")
		c.String(200, ip.(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.1")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if w.Body.String() != "1.2.3.4" {
		t.Fatalf("clientIP: got %q, want %q", w.Body.String(), "1.2.3.4")
	}
}

func TestTrustedProxy_FallbackRemoteAddr(t *testing.T) {
	trusted, _ := security.ParseProxyList([]string{})
	ext := security.NewIPExtractor(trusted)

	r := gin.New()
	r.Use(TrustedProxy(ext))
	r.GET("/test", func(c *gin.Context) {
		ip, _ := c.Get("clientIP")
		c.String(200, ip.(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if w.Body.String() != "192.0.2.1" {
		t.Fatalf("clientIP: got %q, want %q", w.Body.String(), "192.0.2.1")
	}
}

func TestSecurityHeaders_DefaultHeaders(t *testing.T) {
	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "no-referrer",
	}

	r := gin.New()
	r.Use(SecurityHeaders(headers))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("X-Content-Type-Options: got %q, want %q",
			w.Header().Get("X-Content-Type-Options"), "nosniff")
	}
	if w.Header().Get("Referrer-Policy") != "no-referrer" {
		t.Fatalf("Referrer-Policy: got %q, want %q",
			w.Header().Get("Referrer-Policy"), "no-referrer")
	}
}

func TestSecurityHeaders_DoNotOverride(t *testing.T) {
	headers := map[string]string{
		"Referrer-Policy": "no-referrer",
	}

	r := gin.New()
	r.Use(SecurityHeaders(headers))
	r.GET("/test", func(c *gin.Context) {
		c.Header("Referrer-Policy", "origin")
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("Referrer-Policy") != "origin" {
		t.Fatalf("Referrer-Policy: got %q, want %q",
			w.Header().Get("Referrer-Policy"), "origin")
	}
}

func TestValidateSecretKey_ProductionRejectsDefault(t *testing.T) {
	cfg := &config.Config{
		General: config.GeneralConfig{Debug: false},
		Server:  config.ServerConfig{SecretKey: "ultrasecretkey"},
	}
	err := ValidateSecretKey(cfg)
	if err == nil {
		t.Fatal("expected error for default secret key in production")
	}
}

func TestValidateSecretKey_DebugAllowsDefault(t *testing.T) {
	cfg := &config.Config{
		General: config.GeneralConfig{Debug: true},
		Server:  config.ServerConfig{SecretKey: "ultrasecretkey"},
	}
	err := ValidateSecretKey(cfg)
	if err != nil {
		t.Fatalf("debug mode should allow default secret key: %v", err)
	}
}

func TestValidateSecretKey_CustomPasses(t *testing.T) {
	cfg := &config.Config{
		General: config.GeneralConfig{Debug: false},
		Server:  config.ServerConfig{SecretKey: "real-secret-123"},
	}
	err := ValidateSecretKey(cfg)
	if err != nil {
		t.Fatalf("custom secret key should pass: %v", err)
	}
}

func TestRobotsTxt(t *testing.T) {
	r := gin.New()
	r.GET("/robots.txt", HandleRobotsTxt)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/robots.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if len(body) == 0 {
		t.Fatal("empty response body")
	}
}
