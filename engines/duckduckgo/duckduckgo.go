package duckduckgo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine.Register("duckduckgo", &DuckDuckGo{})
}

type DuckDuckGo struct {
	client *httpx.Client
}

func (d *DuckDuckGo) Name() string {
	return "duckduckgo"
}

func (d *DuckDuckGo) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral}
}

func (d *DuckDuckGo) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsPagination: true,
		SupportsTimeRange:  true,
	}
}

// timeRangeParam maps SearXNG-style time range values to DuckDuckGo's df parameter.
func timeRangeParam(tr string) string {
	switch tr {
	case "day":
		return "d"
	case "week":
		return "w"
	case "month":
		return "m"
	case "year":
		return "y"
	}
	return ""
}

func (d *DuckDuckGo) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (d *DuckDuckGo) Setup(cfg engine.EngineInitConfig) bool {
	d.client = cfg.Client
	return true
}

func (d *DuckDuckGo) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://duckduckgo.com",
		WikidataID: "Q1041718",
	}
}

func (d *DuckDuckGo) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	formData := map[string]string{
		"q":  req.Query,
		"kl": "en-us",
	}

	if req.SafeSearch > 0 {
		formData["kp"] = "1"
	}

	if req.Page > 1 {
		formData["s"] = strconv.Itoa((req.Page - 1) * 30)
	}

	if df := timeRangeParam(req.TimeRange); df != "" {
		formData["df"] = df
	}

	resp, err := d.client.R().
		SetContext(ctx).
		SetFormData(formData).
		Post("https://html.duckduckgo.com/html/")
	if err != nil {
		return nil, fmt.Errorf("duckduckgo request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var results []models.Result

	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find(".result__a")
		title := strings.TrimSpace(titleElem.Text())
		href, _ := titleElem.Attr("href")
		snippet := strings.TrimSpace(s.Find(".result__snippet").Text())

		if title != "" && href != "" {
			results = append(results, models.Result{
				Title:    title,
				URL:      href,
				Content:  snippet,
				Engine:   d.Name(),
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
