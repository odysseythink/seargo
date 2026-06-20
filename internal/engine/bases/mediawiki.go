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

// MediaWikiConfig defines the configuration for a MediaWiki API engine.
type MediaWikiConfig struct {
	BaseURL string // e.g., "https://en.wikipedia.org/w/api.php"
}

type mediaWikiEngine struct {
	name       string
	categories []models.Category
	cfg        MediaWikiConfig
	client     *httpx.Client
}

// NewMediaWikiEngine creates a MediaWiki API based engine.
func NewMediaWikiEngine(name string, categories []models.Category, cfg MediaWikiConfig) engine.Engine {
	return &mediaWikiEngine{
		name:       name,
		categories: categories,
		cfg:        cfg,
	}
}

func (e *mediaWikiEngine) Name() string                   { return e.name }
func (e *mediaWikiEngine) Categories() []models.Category   { return e.categories }
func (e *mediaWikiEngine) About() engine.EngineAbout       { return engine.EngineAbout{} }
func (e *mediaWikiEngine) Capabilities() engine.Capabilities { return engine.Capabilities{} }

func (e *mediaWikiEngine) Setup(cfg engine.EngineInitConfig) bool {
	return e.cfg.BaseURL != ""
}

func (e *mediaWikiEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (e *mediaWikiEngine) SetClient(c *httpx.Client) {
	e.client = c
}

func (e *mediaWikiEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	params := url.Values{
		"action":   {"query"},
		"list":     {"search"},
		"srsearch": {req.Query},
		"format":   {"json"},
		"srlimit":  {"10"},
	}

	searchURL := e.cfg.BaseURL + "?" + params.Encode()

	resp, err := e.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("mediawiki %s: request failed: %w", e.name, err)
	}

	var result struct {
		Query struct {
			Search []struct {
				Title   string `json:"title"`
				PageID  int    `json:"pageid"`
				Snippet string `json:"snippet"`
			} `json:"search"`
		} `json:"query"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("mediawiki %s: parse response: %w", e.name, err)
	}

	var results []models.Result
	for _, item := range result.Query.Search {
		pageURL := e.cfg.BaseURL
		if strings.Contains(e.cfg.BaseURL, "/w/api.php") {
			// Wikipedia convention: spaces → underscores, no percent-encoding on parens
			wikiTitle := strings.ReplaceAll(item.Title, " ", "_")
			pageURL = strings.Replace(e.cfg.BaseURL, "/w/api.php", "/wiki/", 1) + wikiTitle
		}

		results = append(results, models.Result{
			Title:    item.Title,
			URL:      pageURL,
			Content:  htmlToText(item.Snippet),
			Engine:   e.name,
			Category: req.Category,
			Template: "default",
		})
	}

	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  results,
	}, nil
}
