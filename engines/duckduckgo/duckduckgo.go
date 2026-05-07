package duckduckgo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine.Register("duckduckgo", &DuckDuckGo{})
}

type DuckDuckGo struct {
	client *resty.Client
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
	}
}

func (d *DuckDuckGo) Init(cfg map[string]any) error {
	d.client = resty.New().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36").
		SetTimeout(10 * time.Second)
	return nil
}

func (d *DuckDuckGo) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(req.Query))

	if req.SafeSearch {
		searchURL += "&kp=1"
	}

	if req.Page > 1 {
		searchURL += "&s=" + strconv.Itoa((req.Page-1)*30)
	}

	resp, err := d.client.R().SetContext(ctx).Get(searchURL)
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
