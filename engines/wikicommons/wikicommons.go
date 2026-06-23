package wikicommons

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	engine.Register("wikicommons", &Wikicommons{})
}

var apiURL = "https://commons.wikimedia.org/w/api.php"

var searchTypes = map[string]string{
	"image": "bitmap|drawing",
	"video": "video",
	"audio": "audio",
	"file":  "multimedia|office|archive|3d",
}

// Wikicommons queries Wikimedia Commons via the MediaWiki API.
type Wikicommons struct {
	client     *httpx.Client
	searchType string
	categories []models.Category
}

func (w *Wikicommons) Name() string { return "wikicommons" }

func (w *Wikicommons) Categories() []models.Category { return w.categories }

func (w *Wikicommons) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://commons.wikimedia.org",
		WikidataID: "Q565",
	}
}

func (w *Wikicommons) Capabilities() engine.Capabilities {
	return engine.Capabilities{SupportsPagination: true}
}

func (w *Wikicommons) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }

func (w *Wikicommons) Setup(cfg engine.EngineInitConfig) bool {
	w.client = cfg.Client
	w.categories = cfg.Categories
	if len(w.categories) == 0 {
		w.categories = []models.Category{models.CategoryFiles}
	}
	if cfg.Extra != nil {
		if s, ok := cfg.Extra["wc_search_type"].(string); ok {
			w.searchType = s
		}
	}
	_, valid := searchTypes[w.searchType]
	return w.client != nil && valid
}

func (w *Wikicommons) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	filetype := searchTypes[w.searchType]
	lang := "en"
	if req.Language != "" {
		lang = strings.SplitN(req.Language, "-", 2)[0]
	}

	args := url.Values{}
	args.Set("format", "json")
	args.Set("uselang", lang)
	args.Set("action", "query")
	args.Set("prop", "info|imageinfo")
	args.Set("generator", "search")
	args.Set("gsrnamespace", "6")
	args.Set("gsrprop", "snippet")
	args.Set("gsrlimit", "10")
	args.Set("gsroffset", fmt.Sprintf("%d", 10*(req.Page-1)))
	args.Set("gsrsearch", fmt.Sprintf("filetype:%s %s", filetype, req.Query))
	args.Set("iiprop", "url|size|mime")
	args.Set("iiurlheight", "180")

	searchURL := apiURL + "?" + args.Encode()
	resp, err := w.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("wikicommons request failed: %w", err)
	}

	var payload struct {
		Query struct {
			Pages map[string]struct {
				Title     string `json:"title"`
				Snippet   string `json:"snippet"`
				Imageinfo []struct {
					DescriptionURL string `json:"descriptionurl"`
					URL            string `json:"url"`
					Mime           string `json:"mime"`
					ThumbURL       string `json:"thumburl"`
					Size           int64  `json:"size"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("wikicommons parse: %w", err)
	}

	var typed []results.Result
	for _, page := range payload.Query.Pages {
		if len(page.Imageinfo) == 0 {
			continue
		}
		ii := page.Imageinfo[0]

		title := strings.TrimPrefix(page.Title, "File:")
		if idx := strings.LastIndex(title, "."); idx > 0 {
			title = title[:idx]
		}

		switch w.searchType {
		case "file":
			typed = append(typed, &results.FileResult{
				BaseResult: results.BaseResult{
					Title:        title,
					URL:          ii.DescriptionURL,
					Content:      page.Snippet,
					Engine:       w.Name(),
					Category:     string(req.Category),
					ThumbnailURL: ii.ThumbURL,
				},
				FileType: ii.Mime,
				FileSize: ii.Size,
				Filename: path.Base(ii.URL),
			})
		}
	}

	return &models.Response{
		Query:        req.Query,
		Category:     req.Category,
		Results:      results.ToAPIResult(typed),
		TypedResults: toAnySlice(typed),
	}, nil
}

func toAnySlice(typed []results.Result) []any {
	out := make([]any, len(typed))
	for i, r := range typed {
		out[i] = r
	}
	return out
}
