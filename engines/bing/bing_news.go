package bing

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	engine.Register("bing_news", &BingNews{})
}

type BingNews struct {
	client *httpx.Client
}

func (b *BingNews) Name() string { return "bing_news" }

func (b *BingNews) Categories() []models.Category {
	return []models.Category{models.CategoryNews}
}

func (b *BingNews) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsPagination: true,
		SupportsTimeRange:  true,
	}
}

func (b *BingNews) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (b *BingNews) Setup(cfg engine.EngineInitConfig) bool {
	b.client = cfg.Client
	return true
}

func (b *BingNews) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://cn.bing.com/news",
		WikidataID: "Q2878637",
	}
}

var bingNewsTimeMap = map[string]string{
	"day":   `interval="4"`,
	"week":  `interval="7"`,
	"month": `interval="9"`,
}

func (b *BingNews) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageIdx := page - 1

	queryParams := url.Values{}
	queryParams.Set("q", req.Query)
	queryParams.Set("InfiniteScroll", "1")
	queryParams.Set("first", strconv.Itoa(pageIdx*10+1))
	queryParams.Set("SFX", strconv.Itoa(pageIdx))
	queryParams.Set("form", "PTFTNR")

	mkt := resolveBingMKT(req)
	if mkt != "" {
		queryParams.Set("mkt", mkt)
	}

	if req.TimeRange != "" {
		queryParams.Set("qft", bingNewsTimeMap[req.TimeRange])
	}

	searchURL := "https://cn.bing.com/news/infinitescrollajax?" + queryParams.Encode()

	rb := b.client.R().
		SetContext(ctx).
		SetHeader("Accept-Language", bingAcceptLanguage(mkt))

	resp, err := rb.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("bing_news request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var typed []results.Result
	doc.Find(`div[class*="newsitem"]`).Each(func(i int, s *goquery.Selection) {
		linkElem := s.Find("a.title")
		href, exists := linkElem.Attr("href")
		if !exists || href == "" {
			return
		}
		title := strings.TrimSpace(linkElem.Text())
		content := strings.TrimSpace(s.Find("div.snippet").Text())

		var metadata []string
		if aria, ok := s.Find("div[class*=\"source\"] span[aria-label]").Attr("aria-label"); ok && aria != "" {
			metadata = append(metadata, strings.TrimSpace(aria))
		}
		if author, ok := linkElem.Attr("data-author"); ok && author != "" {
			metadata = append(metadata, strings.TrimSpace(author))
		}

		thumbnail, _ := s.Find("a.imagelink img").Attr("src")
		thumbnail = strings.TrimSpace(thumbnail)
		if thumbnail != "" && !strings.HasPrefix(thumbnail, "https://cn.bing.com") {
			thumbnail = "https://cn.bing.com/" + strings.TrimLeft(thumbnail, "/")
		}

		extra := map[string]any{}
		if thumbnail != "" {
			extra["thumbnail"] = thumbnail
		}
		if len(metadata) > 0 {
			extra["metadata"] = strings.Join(metadata, " | ")
		}

		typed = append(typed, &results.NewsResult{
			BaseResult: results.BaseResult{
				Title:      title,
				URL:        href,
				Content:    content,
				Engine:     b.Name(),
				Category:   string(models.CategoryNews),
				EngineData: extra,
			},
		})
	})

	raw := make([]any, len(typed))
	for i, r := range typed {
		raw[i] = r
	}
	return &models.Response{
		Query:        req.Query,
		Category:     models.CategoryNews,
		TypedResults: raw,
	}, nil
}
