package wikipedia

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine.Register("wikipedia", &Wikipedia{})
}

type Wikipedia struct {
	client *httpx.Client
}

func (w *Wikipedia) Name() string { return "wikipedia" }

func (w *Wikipedia) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral}
}

func (w *Wikipedia) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
	}
}

func (w *Wikipedia) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (w *Wikipedia) Setup(cfg engine.EngineInitConfig) bool {
	w.client = cfg.Client
	return true
}

func (w *Wikipedia) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://en.wikipedia.org",
		WikidataID: "Q52",
	}
}

func (w *Wikipedia) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := fmt.Sprintf("https://en.wikipedia.org/w/index.php?search=%s", url.QueryEscape(req.Query))

	resp, err := w.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("wikipedia request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var results []models.Result
	// Wikipedia search results
	doc.Find(".mw-search-result").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find(".mw-search-result-heading a")
		title := strings.TrimSpace(titleElem.Text())
		href, _ := titleElem.Attr("href")
		if href != "" && !strings.HasPrefix(href, "http") {
			href = "https://en.wikipedia.org" + href
		}
		snippet := strings.TrimSpace(s.Find(".searchresult").Text())

		if title != "" && href != "" {
			results = append(results, models.Result{
				Title:    title,
				URL:      href,
				Content:  snippet,
				Engine:   w.Name(),
				Category: req.Category,
			})
		}
	})

	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  results,
	}, nil
}
