package bing

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
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
	engine.Register("bing_videos", &BingVideos{})
}

type BingVideos struct {
	client *httpx.Client
}

func (b *BingVideos) Name() string { return "bing_videos" }

func (b *BingVideos) Categories() []models.Category {
	return []models.Category{models.CategoryVideos}
}

func (b *BingVideos) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsPagination: true,
		SupportsTimeRange:  true,
	}
}

func (b *BingVideos) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (b *BingVideos) Setup(cfg engine.EngineInitConfig) bool {
	b.client = cfg.Client
	return true
}

func (b *BingVideos) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://www.bing.com/videos",
		WikidataID: "Q4914152",
	}
}

func (b *BingVideos) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	queryParams := url.Values{}
	queryParams.Set("q", req.Query)
	queryParams.Set("async", "content")
	queryParams.Set("first", strconv.Itoa((page-1)*35+1))
	queryParams.Set("count", "35")

	mkt := resolveBingMKT(req)
	if mkt != "" {
		queryParams.Set("mkt", mkt)
	}

	if req.TimeRange != "" {
		if minutes, ok := bingImageTimeMap[req.TimeRange]; ok {
			queryParams.Set("form", "VRFLTR")
			queryParams.Set("qft", fmt.Sprintf(" filterui:videoage-lt%d", minutes))
		}
	}

	searchURL := "https://cn.bing.com/videos/asyncv2?" + queryParams.Encode()

	rb := b.client.R().
		SetContext(ctx).
		SetHeader("Accept-Language", bingAcceptLanguage(mkt))

	resp, err := rb.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("bing_videos request failed: %w", err)
	}

	body := resp.String()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}
	// Bing video async endpoint sometimes returns a JS wrapper with the real HTML
	// inside a <noscript> tag; unescape and parse that fallback.
	if noscript := doc.Find("noscript").First(); noscript.Length() > 0 {
		inner, err := noscript.Html()
		if err == nil && strings.Contains(inner, "mc_vtvc_video") {
			inner = html.UnescapeString(inner)
			fallback, err := goquery.NewDocumentFromReader(strings.NewReader(inner))
			if err == nil {
				doc = fallback
			}
		}
	}

	var typed []results.Result
	doc.Find(`div[id*="mc_vtvc_video"]`).Each(func(i int, s *goquery.Selection) {
		vrhdata, exists := s.Find("div.vrhdata").Attr("vrhm")
		if !exists || vrhdata == "" {
			return
		}

		var meta struct {
			Murl string `json:"murl"`
			Vt   string `json:"vt"`
			Du   string `json:"du"`
		}
		if err := json.Unmarshal([]byte(vrhdata), &meta); err != nil {
			return
		}
		if meta.Murl == "" {
			return
		}

		thumbnail, _ := s.Find(`img[class^="rms"]`).Attr("data-src-hq")
		info := strings.TrimSpace(s.Find("div.mc_vtvc_meta_block span").Text())

		typed = append(typed, &results.VideoResult{
			BaseResult: results.BaseResult{
				Title:    meta.Vt,
				URL:      meta.Murl,
				Content:  info,
				Engine:   b.Name(),
				Category: string(models.CategoryVideos),
			},
			Thumbnail: thumbnail,
			Length:    meta.Du,
		})
	})

	raw := make([]any, len(typed))
	for i, r := range typed {
		raw[i] = r
	}
	return &models.Response{
		Query:        req.Query,
		Category:     models.CategoryVideos,
		TypedResults: raw,
	}, nil
}
