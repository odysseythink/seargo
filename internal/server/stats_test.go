package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/metrics"
)

func setupStatsTestServer() (*Server, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	store := metrics.NewEngineStatsStore(100)
	store.Record("google", 500*time.Millisecond, 300*time.Millisecond, 10, nil)
	store.Record("google", 600*time.Millisecond, 350*time.Millisecond, 8, nil)
	store.Record("bing", 2*time.Second, 2*time.Second, 0, http.ErrHandlerTimeout)
	store.Record("ddg", 100*time.Millisecond, 80*time.Millisecond, 5, nil)
	store.SetSuspended("bing", true)
	s := &Server{
		router:            gin.New(),
		config:            &config.Config{},
		enginesStatsStore: store,
	}
	return s, s.router
}

func TestHandleStatsEngines(t *testing.T) {
	s, r := setupStatsTestServer()
	r.GET("/api/stats/engines", s.handleStatsEngines)

	req := httptest.NewRequest("GET", "/api/stats/engines", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Engines []metrics.EngineSnapshot `json:"engines"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(resp.Engines) != 3 {
		t.Fatalf("expected 3 engines, got %d", len(resp.Engines))
	}

	var google *metrics.EngineSnapshot
	for i := range resp.Engines {
		if resp.Engines[i].Engine == "google" {
			google = &resp.Engines[i]
			break
		}
	}
	if google == nil {
		t.Fatal("google not found in response")
	}
	if google.RequestCount != 2 {
		t.Errorf("google RequestCount expected 2, got %d", google.RequestCount)
	}
	if google.Reliability != 1.0 {
		t.Errorf("google Reliability expected 1.0, got %v", google.Reliability)
	}

	var bing *metrics.EngineSnapshot
	for i := range resp.Engines {
		if resp.Engines[i].Engine == "bing" {
			bing = &resp.Engines[i]
			break
		}
	}
	if bing == nil {
		t.Fatal("bing not found in response")
	}
	if !bing.Suspended {
		t.Error("bing should be suspended")
	}
	if bing.Reliability != 0.0 {
		t.Errorf("bing Reliability expected 0.0, got %v", bing.Reliability)
	}
}

func TestHandleStatsErrors(t *testing.T) {
	s, r := setupStatsTestServer()
	r.GET("/api/stats/errors", s.handleStatsErrors)

	req := httptest.NewRequest("GET", "/api/stats/errors", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Errors []map[string]any `json:"errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(resp.Errors) == 0 {
		t.Fatal("expected at least 1 engine with errors")
	}

	foundBing := false
	for _, e := range resp.Errors {
		if e["engine"] == "bing" {
			foundBing = true
			if e["total_errors"] == nil || e["total_errors"].(float64) == 0 {
				t.Error("bing should have errors")
			}
			break
		}
	}
	if !foundBing {
		t.Error("bing not found in errors response")
	}
}

func TestHandleStatsEnginesEmpty(t *testing.T) {
	store := metrics.NewEngineStatsStore(100)
	s := &Server{
		router:            gin.New(),
		config:            &config.Config{},
		enginesStatsStore: store,
	}
	s.router.GET("/api/stats/engines", s.handleStatsEngines)

	req := httptest.NewRequest("GET", "/api/stats/engines", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 even when empty, got %d", w.Code)
	}
	var resp struct {
		Engines []metrics.EngineSnapshot `json:"engines"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Engines) != 0 {
		t.Errorf("expected empty engines list, got %d", len(resp.Engines))
	}
}
