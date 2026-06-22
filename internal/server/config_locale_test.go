package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/i18n"
)

func TestHandleConfig_ReturnsRTL(t *testing.T) {
	cfg := minimalTestConfig()
	cfg.UI.DefaultLocale = "ar" // Arabic is RTL
	reg := i18n.NewLocaleRegistry()
	s := &Server{router: gin.New(), config: cfg, localeRegistry: reg}

	gin.SetMode(gin.TestMode)
	s.router.GET("/api/config", s.handleConfig)

	req := httptest.NewRequest("GET", "/api/config", nil)
	req.Header.Set("Accept-Language", "ar")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	ui := resp["ui"].(map[string]interface{})
	rtl, ok := ui["rtl"].(bool)
	if !ok {
		t.Fatal("ui.rtl missing or not bool")
	}
	if !rtl {
		t.Error("ar locale should set rtl=true")
	}
}

func TestHandleConfig_ReturnsDefaultLocale(t *testing.T) {
	cfg := minimalTestConfig()
	cfg.UI.DefaultLocale = "zh-CN"
	reg := i18n.NewLocaleRegistry()
	s := &Server{router: gin.New(), config: cfg, localeRegistry: reg}

	gin.SetMode(gin.TestMode)
	s.router.GET("/api/config", s.handleConfig)

	// No Accept-Language header → negotiator falls back to config default (zh-CN)
	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	ui := resp["ui"].(map[string]interface{})
	if ui["default_locale"] != "zh-CN" {
		t.Errorf("default_locale = %v, want zh-CN", ui["default_locale"])
	}
}
