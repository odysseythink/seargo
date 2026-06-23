package bases

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

// XPathConfig defines the extraction rules for an xpath-based engine.
type XPathConfig struct {
	SearchURL    string // URL template with {query}, {page}, {lang} placeholders
	ResultXPath  string // XPath to select individual result containers
	URLXPath     string // XPath for result URL (relative to ResultXPath)
	TitleXPath   string // XPath for result title
	ContentXPath string // XPath for result content/snippet
	ThumbnailXPath string // XPath for thumbnail URL (optional)

	// Pagination
	Paging      bool   // whether engine supports pagination
	PageSize    int    // results per page
	FirstPage   int    // first page number (usually 0 or 1)
	PageField   string // query param name for page

	// Language
	LanguageSupport bool
	LanguageParam   string // query param name for language

	// Typed result support
	ResultType ResultTypeConfig
}

// xpathEngine implements engine.Engine using XPath-based HTML scraping.
type xpathEngine struct {
	name       string
	categories []models.Category
	cfg        XPathConfig
	client     *httpx.Client
}

// NewXPathEngine creates an xpath-based engine from config.
func NewXPathEngine(name string, categories []models.Category, cfg XPathConfig) engine.Engine {
	if cfg.PageSize <= 0 {
		cfg.PageSize = 10
	}
	if cfg.FirstPage == 0 {
		cfg.FirstPage = 1
	}
	return &xpathEngine{
		name:       name,
		categories: categories,
		cfg:        cfg,
	}
}

func (e *xpathEngine) Name() string                 { return e.name }
func (e *xpathEngine) Categories() []models.Category { return e.categories }
func (e *xpathEngine) About() engine.EngineAbout     { return engine.EngineAbout{} }

func (e *xpathEngine) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsPagination: e.cfg.Paging,
		SupportsLanguage:   e.cfg.LanguageSupport,
	}
}

func (e *xpathEngine) Setup(cfg engine.EngineInitConfig) bool {
	if e.cfg.SearchURL == "" {
		return false
	}
	if e.cfg.ResultXPath == "" {
		return false
	}
	e.client = cfg.Client
	if e.cfg.ResultType.Type == "" {
		e.cfg.ResultType = parseResultTypeConfig(cfg.Extra)
	}
	return true
}

func (e *xpathEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

// SetClient sets the HTTP client for this engine (called by scheduler).
func (e *xpathEngine) SetClient(c *httpx.Client) {
	e.client = c
}

func (e *xpathEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := e.buildURL(req)

	resp, err := e.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("xpath engine %s: request failed: %w", e.name, err)
	}

	doc, err := htmlquery.Parse(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("xpath engine %s: parse HTML: %w", e.name, err)
	}

	resultNodes, err := htmlquery.QueryAll(doc, e.cfg.ResultXPath)
	if err != nil {
		return nil, fmt.Errorf("xpath engine %s: query results: %w", e.name, err)
	}

	var typed []results.Result
	var apiResults []models.Result
	for _, node := range resultNodes {
		base := models.Result{
			Engine:   e.name,
			Category: req.Category,
			Template: "default",
		}

		if e.cfg.URLXPath != "" {
			base.URL = evalXPathAttrOne(node, e.cfg.URLXPath, searchURL)
		}
		if e.cfg.TitleXPath != "" {
			base.Title = evalXPathGetOne(node, e.cfg.TitleXPath)
		}
		if e.cfg.ContentXPath != "" {
			base.Content = evalXPathGetOne(node, e.cfg.ContentXPath)
		}
		if e.cfg.ThumbnailXPath != "" {
			base.ThumbnailURL = evalXPathAttrOne(node, e.cfg.ThumbnailXPath, searchURL)
		}

		if base.URL == "" || base.Title == "" {
			continue
		}

		if e.cfg.ResultType.Type != "" {
			tr := buildTypedResult(node, e.cfg.ResultType, base)
			typed = append(typed, tr)
			apiResults = append(apiResults, results.ToAPIResult([]results.Result{tr})...)
		} else {
			apiResults = append(apiResults, base)
		}
	}

	var raw []any
	if len(typed) > 0 {
		raw = make([]any, len(typed))
		for i, r := range typed {
			raw[i] = r
		}
	}

	return &models.Response{
		Query:        req.Query,
		Category:     req.Category,
		Results:      apiResults,
		TypedResults: raw,
	}, nil
}

// buildURL constructs the search URL by substituting placeholders.
func (e *xpathEngine) buildURL(req *models.Request) string {
	u := e.cfg.SearchURL
	u = strings.ReplaceAll(u, "{query}", url.QueryEscape(req.Query))
	if req.Language != "" {
		u = strings.ReplaceAll(u, "{lang}", url.QueryEscape(req.Language))
	}
	if e.cfg.Paging {
		page := req.Page
		if page < e.cfg.FirstPage {
			page = e.cfg.FirstPage
		}
		u = strings.ReplaceAll(u, "{page}", fmt.Sprintf("%d", page))
	}
	return u
}

// evalXPathAttrOne evaluates an XPath expression and returns the attribute
// value from the first matching node. If the value is a relative URL, it is
// resolved against baseURL.
func evalXPathAttrOne(node *html.Node, xpath, baseURL string) string {
	found, err := htmlquery.Query(node, xpath)
	if err != nil || found == nil {
		return ""
	}
	val := strings.TrimSpace(htmlquery.InnerText(found))
	if val == "" {
		return ""
	}
	return extractURL(baseURL, val)
}
