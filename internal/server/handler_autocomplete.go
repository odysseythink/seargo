package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/metrics"
)

// autocompleteRequest binds query parameters for /api/autocomplete.
type autocompleteRequest struct {
	Query   string `form:"q"`
	Backend string `form:"backend"`
	Format  string `form:"format"` // "json" (default) or "x-suggestions+json"
}

// autocompleteSuggestion is a single suggestion in the API response.
type autocompleteSuggestion struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// autocompleteResponse is the JSON response for /api/autocomplete.
type autocompleteResponse struct {
	Query       string                   `json:"query"`
	Suggestions []autocompleteSuggestion `json:"suggestions"`
}

func (s *Server) handleAutocomplete(c *gin.Context) {
	var req autocompleteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing q parameter"})
		return
	}

	// Rate limit
	clientIP := c.ClientIP()
	if s.rateLimiter != nil && !s.rateLimiter.Allow(clientIP) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}

	// Determine backend: user override > config default
	backend := req.Backend
	if backend == "" {
		backend = s.config.Search.Autocomplete
	}
	if backend == "" {
		backend = "google"
	}

	start := time.Now()
	defer func() {
		metrics.AutocompleteDurationSeconds.WithLabelValues(backend).Observe(time.Since(start).Seconds())
	}()

	// Check for bang prefix — delegate to bangs service
	isBang := strings.HasPrefix(req.Query, "!!") || strings.HasPrefix(req.Query, "!")
	var suggestions []string

	if isBang {
		prefix := strings.TrimLeft(req.Query, "!")
		if s.bangsService != nil {
			suggestions = s.bangsService.Suggest(prefix)
			bangPrefix := "!!"
			if strings.HasPrefix(req.Query, "!") && !strings.HasPrefix(req.Query, "!!") {
				bangPrefix = "!"
			}
			for i := range suggestions {
				suggestions[i] = bangPrefix + suggestions[i]
			}
		}
	} else if s.autocomplete != nil {
		locale := s.config.Search.DefaultLang
		if locale == "" {
			locale = "en"
		}
		suggestions = s.autocomplete.Suggest(c.Request.Context(), backend, req.Query, locale)
	}

	metrics.AutocompleteRequestsTotal.WithLabelValues(backend).Inc()

	var result []autocompleteSuggestion
	for _, sug := range suggestions {
		if sug != "" {
			result = append(result, autocompleteSuggestion{Label: sug, Value: sug})
		}
	}

	if req.Format == "x-suggestions+json" {
		values := make([]string, 0, len(result))
		for _, s := range result {
			values = append(values, s.Value)
		}
		c.JSON(http.StatusOK, []interface{}{req.Query, values})
		return
	}

	c.JSON(http.StatusOK, autocompleteResponse{
		Query:       req.Query,
		Suggestions: result,
	})
}

func (s *Server) handleOpenSearch(c *gin.Context) {
	baseURL := "http://" + c.Request.Host
	if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
		baseURL = "https://" + c.Request.Host
	}

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
  <ShortName>SearGo</ShortName>
  <Description>Search with SearGo</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <Image width="16" height="16" type="image/x-icon">` + baseURL + `/favicon.ico</Image>
  <Url type="text/html" template="` + baseURL + `/search?q={searchTerms}"/>
  <Url type="application/x-suggestions+json" template="` + baseURL + `/api/autocomplete?q={searchTerms}&amp;format=x-suggestions%2Bjson"/>
  <Url type="application/opensearchdescription+xml" rel="self" template="` + baseURL + `/api/opensearch.xml"/>
</OpenSearchDescription>`

	c.Data(http.StatusOK, "application/opensearchdescription+xml", []byte(xml))
}
