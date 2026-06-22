package yahoo

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
	engine.Register("yahoo", &Yahoo{})
}

type Yahoo struct {
	client *httpx.Client
}

func (y *Yahoo) Name() string { return "yahoo" }

func (y *Yahoo) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral}
}

func (y *Yahoo) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsPagination: true,
	}
}

func (y *Yahoo) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (y *Yahoo) Setup(cfg engine.EngineInitConfig) bool {
	y.client = cfg.Client
	return true
}

func (y *Yahoo) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://search.yahoo.com",
		WikidataID: "Q37081",
	}
}

func (y *Yahoo) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := fmt.Sprintf("https://search.yahoo.com/search?p=%s", url.QueryEscape(req.Query))

	resp, err := y.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("yahoo request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var results []models.Result
	doc.Find(".algo").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find("h3 a, h4 a, .title a")
		title := strings.TrimSpace(titleElem.Text())
		href, _ := titleElem.Attr("href")
		snippet := strings.TrimSpace(s.Find(".fc-1st, .abstract").Text())

		if title != "" && href != "" {
			results = append(results, models.Result{
				Title:    title,
				URL:      href,
				Content:  snippet,
				Engine:   y.Name(),
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
