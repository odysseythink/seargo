package bing

import (
	"context"
	"encoding/json"
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
	engine.Register("bing_images", &BingImages{})
}

type BingImages struct {
	client *httpx.Client
}

func (b *BingImages) Name() string { return "bing_images" }

func (b *BingImages) Categories() []models.Category {
	return []models.Category{models.CategoryImages}
}

func (b *BingImages) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsPagination: true,
		SupportsTimeRange:  true,
	}
}

func (b *BingImages) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (b *BingImages) Setup(cfg engine.EngineInitConfig) bool {
	b.client = cfg.Client
	return true
}

func (b *BingImages) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://www.bing.com/images",
		WikidataID: "Q182496",
	}
}

var bingImageTimeMap = map[string]int{
	"day":   60 * 24,
	"week":  60 * 24 * 7,
	"month": 60 * 24 * 31,
	"year":  60 * 24 * 365,
}

func (b *BingImages) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	queryParams := url.Values{}
	queryParams.Set("q", req.Query)
	queryParams.Set("async", "1")
	queryParams.Set("first", strconv.Itoa((page-1)*35+1))
	queryParams.Set("count", "35")

	mkt := resolveBingMKT(req)
	if mkt != "" {
		queryParams.Set("mkt", mkt)
	}

	if req.TimeRange != "" {
		if minutes, ok := bingImageTimeMap[req.TimeRange]; ok {
			queryParams.Set("qft", fmt.Sprintf("filterui:age-lt%d", minutes))
		}
	}

	searchURL := "https://cn.bing.com/images/async?" + queryParams.Encode()

	rb := b.client.R().
		SetContext(ctx).
		SetHeader("Accept-Language", bingAcceptLanguage(mkt))

	resp, err := rb.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("bing_images request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var typed []results.Result
	doc.Find(`ul[class*="dgControl_list"] > li`).Each(func(i int, s *goquery.Selection) {
		metaAttr, exists := s.Find("a.iusc").Attr("m")
		if !exists || metaAttr == "" {
			return
		}

		var meta struct {
			Purl string `json:"purl"`
			Turl string `json:"turl"`
			Murl string `json:"murl"`
			Desc string `json:"desc"`
		}
		if err := json.Unmarshal([]byte(metaAttr), &meta); err != nil {
			return
		}
		if meta.Purl == "" {
			return
		}

		title := strings.TrimSpace(s.Find("div.infnmpt a").Text())
		if title == "" {
			title = req.Query
		}

		formatParts := strings.Split(strings.TrimSpace(s.Find("div.imgpt div span").Text()), " · ")
		source := strings.TrimSpace(s.Find("div.imgpt div.lnkw a").Text())

		typed = append(typed, &results.ImageResult{
			BaseResult: results.BaseResult{
				Title:    title,
				URL:      meta.Purl,
				Content:  meta.Desc,
				Engine:   b.Name(),
				Category: string(models.CategoryImages),
			},
			ImgSrc:       meta.Murl,
			ThumbnailSrc: meta.Turl,
			Source:       source,
			Resolution:   firstOrEmpty(formatParts, 0),
			ImgFormat:    firstOrEmpty(formatParts, 1),
		})
	})

	raw := make([]any, len(typed))
	for i, r := range typed {
		raw[i] = r
	}
	return &models.Response{
		Query:        req.Query,
		Category:     models.CategoryImages,
		TypedResults: raw,
	}, nil
}

func firstOrEmpty(parts []string, idx int) string {
	if idx >= 0 && idx < len(parts) {
		return strings.TrimSpace(parts[idx])
	}
	return ""
}
