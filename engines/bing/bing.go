package bing

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine.Register("bing", &Bing{})
}

type Bing struct {
	client *resty.Client
}

func (b *Bing) Name() string { return "bing" }

func (b *Bing) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral}
}

func (b *Bing) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsPagination: true,
	}
}

func (b *Bing) Init(cfg map[string]any) error {
	b.client = resty.New().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36").
		SetTimeout(10 * time.Second)
	return nil
}

func (b *Bing) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(req.Query))

	resp, err := b.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("bing request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var results []models.Result
	doc.Find(".b_algo").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find("h2 a")
		title := strings.TrimSpace(titleElem.Text())
		href, _ := titleElem.Attr("href")
		snippet := strings.TrimSpace(s.Find(".b_caption p").Text())
		if snippet == "" {
			snippet = strings.TrimSpace(s.Find("p").Text())
		}

		if title != "" && href != "" {
			results = append(results, models.Result{
				Title:    title,
				URL:      href,
				Content:  snippet,
				Engine:   b.Name(),
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
