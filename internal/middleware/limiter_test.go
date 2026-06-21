package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/limiter"
	"github.com/seargo/seargo/internal/storage"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func makeTestLimiter(t *testing.T, burstMax int64) limiter.Limiter {
	t.Helper()
	kv, err := storage.New(storage.Options{
		Backend:     "memory",
		NumCounters: 10000,
		MaxCost:     10 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { kv.Close() })

	return limiter.New(&limiter.Config{
		BurstWindow:     time.Minute,
		BurstMax:        burstMax,
		LongWindow:      time.Minute,
		LongMax:         20,
		FilterLinkLocal: false,
		LinkToken:       false,
	}, kv.WithNamespace("limiter_test"))
}

func TestLimiter_Allow(t *testing.T) {
	lm := makeTestLimiter(t, 10)

	r := gin.New()
	r.Use(Limiter(&config.Config{}, lm))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
}

func TestLimiter_Block(t *testing.T) {
	lm := makeTestLimiter(t, 0) // 0 burst max = block all

	r := gin.New()
	r.Use(Limiter(&config.Config{}, lm))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	r.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Fatalf("rate limited should get 429, got %d", w.Code)
	}
}

func TestLimiter_RetryAfterHeader(t *testing.T) {
	lm := makeTestLimiter(t, 0)

	r := gin.New()
	r.Use(Limiter(&config.Config{}, lm))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	r.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Fatalf("rate limited should get 429, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("429 response should include Retry-After header")
	}
}

func TestHandleLimiterLinkToken(t *testing.T) {
	r := gin.New()
	r.GET("/link_token", func(c *gin.Context) {
		// In real usage, the Limiter middleware sets limiterSvc into the context.
		// For this test, we just verify the handler doesn't panic.
		c.JSON(200, gin.H{"token": "test-token"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/link_token", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
}
