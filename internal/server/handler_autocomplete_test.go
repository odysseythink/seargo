package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/config"
)

func TestHandleAutocomplete_MissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Search: config.SearchConfig{
			Autocomplete:    "google",
			AutocompleteMin: 4,
			DefaultLang:     "en",
		},
	}

	srv := New(cfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := srv.router

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/autocomplete", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "missing q parameter", body["error"])
}

func TestHandleAutocomplete_NilAutocompleteReturnsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Search: config.SearchConfig{
			Autocomplete:    "google",
			AutocompleteMin: 4,
			DefaultLang:     "en",
		},
	}

	srv := New(cfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := srv.router

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/autocomplete?q=test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp autocompleteResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "test", resp.Query)
	assert.Empty(t, resp.Suggestions)
}

func TestHandleAutocomplete_RateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Search: config.SearchConfig{
			Autocomplete: "google",
			DefaultLang:  "en",
		},
	}
	rl := NewRateLimiter(1, time.Hour) // 1 token per hour
	defer rl.Close()

	srv := New(cfg, nil, nil, nil, rl, nil, nil, nil, nil, nil, nil, nil)
	router := srv.router

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/autocomplete?q=test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/autocomplete?q=test2", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestHandleOpenSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Search: config.SearchConfig{DefaultLang: "en"},
	}
	srv := New(cfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := srv.router

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/opensearch.xml", nil)
	req.Host = "localhost:8080"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "<OpenSearchDescription")
	assert.Contains(t, w.Body.String(), "SearGo")
	assert.Contains(t, w.Body.String(), "localhost:8080")
}
