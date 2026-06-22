package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, exists := c.Get("request_id")
		if !exists {
			t.Error("request_id not found in context")
		}
		if id.(string) == "" {
			t.Error("request_id is empty")
		}
		c.String(200, "ok")
	})
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	respID := w.Header().Get("X-Request-ID")
	if respID == "" {
		t.Error("X-Request-ID response header missing")
	}
}

func TestRequestIDPreservesValidIncoming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("request_id")
		if id.(string) != "abc-123-def" {
			t.Errorf("expected 'abc-123-def', got %q", id)
		}
		c.String(200, "ok")
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "abc-123-def")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	respID := w.Header().Get("X-Request-ID")
	if respID != "abc-123-def" {
		t.Errorf("expected 'abc-123-def' in response, got %q", respID)
	}
}

func TestRequestIDTruncatesTooLong(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("request_id")
		idStr := id.(string)
		if len(idStr) > 64 {
			t.Errorf("request_id too long: %d chars", len(idStr))
		}
		if !strings.HasPrefix(idStr, "a") {
			t.Error("truncated request_id should start with beginning of header")
		}
		c.String(200, "ok")
	})
	longID := strings.Repeat("a", 100)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", longID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	respID := w.Header().Get("X-Request-ID")
	if len(respID) > 64 {
		t.Errorf("response X-Request-ID too long: %d", len(respID))
	}
}

func TestRequestIDRejectsInvalidChars(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("request_id")
		idStr := id.(string)
		if strings.Contains(idStr, "<") || strings.Contains(idStr, ">") {
			t.Errorf("request_id contains invalid chars: %q", idStr)
		}
		c.String(200, "ok")
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "<script>alert(1)</script>")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	respID := w.Header().Get("X-Request-ID")
	if respID == "" {
		t.Error("X-Request-ID should be generated")
	}
	if respID == "<script>alert(1)</script>" {
		t.Error("invalid request_id should not be echoed")
	}
}

func TestRequestIDEmptyHeaderGenerates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("request_id")
		if id.(string) == "" {
			t.Error("request_id should not be empty")
		}
		c.String(200, "ok")
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	respID := w.Header().Get("X-Request-ID")
	if respID == "" {
		t.Error("X-Request-ID should be generated for empty header")
	}
}
