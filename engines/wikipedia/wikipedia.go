package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/seargo/seargo/engines/wikimedia"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	engine.Register("wikipedia", &Wikipedia{})
}

var restSummaryURL = "https://%s/api/rest_v1/page/summary/%s"

type Wikipedia struct {
	client     *httpx.Client
	traits     engine.EngineTraits
	wikiNetloc map[string]string
}

func (w *Wikipedia) Name() string { return "wikipedia" }

func (w *Wikipedia) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral}
}

func (w *Wikipedia) Capabilities() engine.Capabilities {
	return engine.Capabilities{SupportsLanguage: true}
}

func (w *Wikipedia) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://en.wikipedia.org",
		WikidataID: "Q52",
	}
}

func (w *Wikipedia) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	store := wikimedia.NewWikiNetlocStore(cfg.Client, wikiNetlocCachePath())
	mapping, ok := store.LoadOrFetch(ctx)
	if ok {
		w.wikiNetloc = mapping
	}
	// Non-fatal: if the netloc fetch/cache fails, ResolveWikiNetloc falls back
	// to <lang>.wikipedia.org, so the engine can still serve requests.
	return true
}

func (w *Wikipedia) Setup(cfg engine.EngineInitConfig) bool {
	w.client = cfg.Client
	w.traits = cfg.EngineTraits
	return true
}

func (w *Wikipedia) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	if w.client == nil {
		return emptyResponse(req), nil
	}

	query := req.Query
	// Title-case single-word queries for better REST API matching, but only when
	// the input is all lower-case.  strings.Title is deprecated but kept here
	// because importing golang.org/x/text for a single call adds unacceptable
	// dependency weight for this project.
	if query == strings.ToLower(query) && !strings.Contains(query, " ") {
		query = strings.ToUpper(query[:1]) + query[1:]
	}

	_, netloc := wikimedia.ResolveWikiNetloc(w.traits, w.wikiNetloc, req.Language)
	summaryURL := fmt.Sprintf(restSummaryURL, netloc, url.PathEscape(query))

	resp, err := w.client.R().SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(summaryURL)
	if err != nil && resp == nil {
		return nil, fmt.Errorf("wikipedia request failed: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("wikipedia request failed: %w", err)
	}

	if resp.StatusCode == 404 {
		return emptyResponse(req), nil
	}
	if resp.StatusCode == 400 {
		var apiErr wikiErrorResponse
		if json.Unmarshal(resp.Body, &apiErr) == nil &&
			apiErr.Type == "https://mediawiki.org/wiki/HyperSwitch/errors/bad_request" &&
			apiErr.Detail == "title-invalid-characters" {
			return emptyResponse(req), nil
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("wikipedia unexpected status %d", resp.StatusCode)
	}

	var api wikiSummaryResponse
	if err := json.Unmarshal(resp.Body, &api); err != nil {
		return nil, fmt.Errorf("wikipedia parse: %w", err)
	}

	title := wikimedia.HTMLToText(coalesce(api.Titles.Display, api.Title))
	pageURL := api.ContentURLs.Desktop.Page

	if api.Type != "standard" {
		typed := []results.Result{
			&results.MainResult{
				BaseResult: results.BaseResult{
					Title:    title,
					URL:      pageURL,
					Content:  api.Description,
					Engine:   w.Name(),
					Template: "default.html",
					Category: string(req.Category),
				},
			},
		}
		return responseFromTyped(req, typed), nil
	}

	infobox := &results.InfoboxResult{
		BaseResult: results.BaseResult{
			Title:    title,
			Content:  api.Extract,
			Engine:   w.Name(),
			Template: "infobox.html",
			Category: string(req.Category),
		},
		InfoboxID: pageURL,
		ImgAlt:    title,
	}
	if api.Thumbnail.Source != "" {
		infobox.ImgSrc = api.Thumbnail.Source
	}
	infobox.URLs = []results.InfoboxURL{{Title: "Wikipedia", URL: pageURL}}

	return responseFromTyped(req, []results.Result{infobox}), nil
}

type wikiSummaryResponse struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Titles      struct {
		Display string `json:"display"`
	} `json:"titles"`
	Extract     string `json:"extract"`
	Description string `json:"description"`
	ContentURLs struct {
		Desktop struct {
			Page string `json:"page"`
		} `json:"desktop"`
	} `json:"content_urls"`
	Thumbnail struct {
		Source string `json:"source"`
	} `json:"thumbnail"`
}

type wikiErrorResponse struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

func wikiNetlocCachePath() string {
	if p := os.Getenv("SEARGO_WIKI_NETLOC_CACHE"); p != "" {
		return p
	}
	return "data/wiki_netloc.json"
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func emptyResponse(req *models.Request) *models.Response {
	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
	}
}

func responseFromTyped(req *models.Request, typed []results.Result) *models.Response {
	raw := make([]any, len(typed))
	for i, r := range typed {
		raw[i] = r
	}
	return &models.Response{
		Query:        req.Query,
		Category:     req.Category,
		Results:      results.ToAPIResult(typed),
		TypedResults: raw,
	}
}
