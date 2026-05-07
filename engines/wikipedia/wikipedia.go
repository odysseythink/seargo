package wikipedia

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
	engine.Register("wikipedia", &Wikipedia{})
}

type Wikipedia struct {
	client *resty.Client
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

func (w *Wikipedia) Init(cfg map[string]any) error {
	w.client = resty.New().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		SetTimeout(8 * time.Second).
		SetRetryCount(1)
	return nil
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
