package bases

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

// JSONEngineConfig defines the extraction rules for a JSON-API-based engine.
type JSONEngineConfig struct {
	SearchURL    string // URL template with {query}, {page}, {lang} placeholders
	ResultsQuery string // slash-delimited path to the results array
	URLQuery     string // relative path within each result item for URL
	TitleQuery   string // relative path within each result item for title
	ContentQuery string // relative path within each result item for content

	// Pagination
	Paging    bool
	PageSize  int
	PageField string

	// Language
	LanguageSupport bool
	LanguageParam   string
}

type jsonEngine struct {
	name       string
	categories []models.Category
	cfg        JSONEngineConfig
	client     *httpx.Client
}

// NewJSONEngine creates a JSON-API-based engine from config.
func NewJSONEngine(name string, categories []models.Category, cfg JSONEngineConfig) engine.Engine {
	return &jsonEngine{
		name:       name,
		categories: categories,
		cfg:        cfg,
	}
}

func (e *jsonEngine) Name() string                 { return e.name }
func (e *jsonEngine) Categories() []models.Category { return e.categories }
func (e *jsonEngine) About() engine.EngineAbout     { return engine.EngineAbout{} }

func (e *jsonEngine) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsPagination: e.cfg.Paging,
		SupportsLanguage:   e.cfg.LanguageSupport,
	}
}

func (e *jsonEngine) Setup(cfg engine.EngineInitConfig) bool {
	if e.cfg.SearchURL == "" {
		return false
	}
	if e.cfg.ResultsQuery == "" {
		return false
	}
	return true
}

func (e *jsonEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (e *jsonEngine) SetClient(c *httpx.Client) {
	e.client = c
}

func (e *jsonEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := e.buildURL(req)

	resp, err := e.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("json engine %s: request failed: %w", e.name, err)
	}

	var data interface{}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, fmt.Errorf("json engine %s: parse JSON: %w", e.name, err)
	}

	resultsList := jsonQuery(data, e.cfg.ResultsQuery)

	var results []models.Result
	for _, item := range resultsList {
		resultItem, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		urlVal := firstString(jsonQuery(resultItem, e.cfg.URLQuery))
		titleVal := firstString(jsonQuery(resultItem, e.cfg.TitleQuery))
		contentVal := firstString(jsonQuery(resultItem, e.cfg.ContentQuery))

		if urlVal != "" && titleVal != "" {
			results = append(results, models.Result{
				Title:    titleVal,
				URL:      urlVal,
				Content:  contentVal,
				Engine:   e.name,
				Category: req.Category,
				Template: "default",
			})
		}
	}

	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  results,
	}, nil
}

func (e *jsonEngine) buildURL(req *models.Request) string {
	u := e.cfg.SearchURL
	u = strings.ReplaceAll(u, "{query}", url.QueryEscape(req.Query))
	if req.Language != "" && e.cfg.LanguageSupport {
		u = strings.ReplaceAll(u, "{lang}", url.QueryEscape(req.Language))
	}
	if e.cfg.Paging {
		u = strings.ReplaceAll(u, "{page}", fmt.Sprintf("%d", req.Page))
	}
	return u
}

func firstString(vals []interface{}) string {
	for _, v := range vals {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}
