package server

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/logger"
)

func TestMain(m *testing.M) {
	flag.Set("logtostderr", "true")
	logger.Init("info", "stdout")
	os.Exit(m.Run())
}

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, BindAddress: "0.0.0.0"},
		Search: config.SearchConfig{DefaultLang: "zh-CN"},
	}

	srv := New(cfg)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	srv.router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestConfigEndpoint(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Search: config.SearchConfig{DefaultLang: "zh-CN", DefaultCategory: "general"},
	}

	srv := New(cfg)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config", nil)
	srv.router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "zh-CN")
}
