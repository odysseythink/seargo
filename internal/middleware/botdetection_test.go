package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seargo/seargo/internal/botdetection"
	"github.com/seargo/seargo/internal/security"
)

func TestBotDetection_Allow(t *testing.T) {
	cfg := &botdetection.Config{}
	det := botdetection.NewDetector(cfg, nil)
	extractor := security.NewIPExtractor(nil)

	r := gin.New()
	r.Use(TrustedProxy(extractor))
	r.Use(BotDetection(det))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("User-Agent", "Mozilla/5.0")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
}

func TestBotDetection_Block(t *testing.T) {
	cfg := &botdetection.Config{}
	det := botdetection.NewDetector(cfg, nil)
	extractor := security.NewIPExtractor(nil)

	r := gin.New()
	r.Use(TrustedProxy(extractor))
	r.Use(BotDetection(det))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("User-Agent", "curl/7.64.1")
	r.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Fatalf("bot should get 429, got %d", w.Code)
	}
}

func TestBotDetection_ExemptHealth(t *testing.T) {
	cfg := &botdetection.Config{}
	det := botdetection.NewDetector(cfg, nil)
	extractor := security.NewIPExtractor(nil)

	r := gin.New()
	r.Use(TrustedProxy(extractor))
	r.Use(BotDetection(det))
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("User-Agent", "curl/7.64.1")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("health should be exempt from bot detection, got %d", w.Code)
	}
}
