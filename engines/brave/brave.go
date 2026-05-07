package brave

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
	engine.Register("brave", &Brave{})
}

type Brave struct {
	client *resty.Client
}

func (b *Brave) Name() string { return "brave" }

func (b *Brave) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral}
}

func (b *Brave) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsPagination: true,
	}
}

func (b *Brave) Init(cfg map[string]any) error {
	b.client = resty.New().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		SetHeader("Referer", "https://search.brave.com/").
		SetTimeout(8 * time.Second).
		SetRetryCount(1)
	return nil
}

func (b *Brave) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := fmt.Sprintf("https://search.brave.com/search?q=%s", url.QueryEscape(req.Query))

	resp, err := b.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("brave request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var results []models.Result
	doc.Find(".snippet").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find("a")
		title := strings.TrimSpace(titleElem.Text())
		href, _ := titleElem.Attr("href")
		snippet := strings.TrimSpace(s.Find(".snippet-description").Text())

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
