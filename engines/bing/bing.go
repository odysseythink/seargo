package bing

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
	engine.Register("bing", &Bing{})
}

type Bing struct {
	client *httpx.Client
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

func (b *Bing) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (b *Bing) Setup(cfg engine.EngineInitConfig) bool {
	return true
}

func (b *Bing) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://www.bing.com",
		WikidataID: "Q182496",
	}
}

func (b *Bing) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(req.Query))

	resp, err := b.client.R().
		SetContext(ctx).
		SetHeader("Cookie", "MUID=; MUIDB=; SRCHD=AF=NOFORM; SRCHUID=V=2&GUID=; SRCHUSR=DOB=20240101").
		Get(searchURL)
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
