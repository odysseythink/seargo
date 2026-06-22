package google

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
	engine.Register("google", &Google{})
}

type Google struct {
	client *httpx.Client
}

func (g *Google) Name() string { return "google" }

func (g *Google) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral, models.CategoryImages}
}

func (g *Google) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsPagination: true,
	}
}

func (g *Google) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (g *Google) Setup(cfg engine.EngineInitConfig) bool {
	g.client = cfg.Client
	return true
}

func (g *Google) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://www.google.com",
		WikidataID: "Q95",
	}
}

func (g *Google) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&hl=en", url.QueryEscape(req.Query))

	resp, err := g.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("google request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var results []models.Result
	selectors := []string{"div.g", "div.srg div.g", "#search div.g"}
	for _, sel := range selectors {
		doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			titleElem := s.Find("h3")
			if titleElem.Length() == 0 {
				titleElem = s.Find("a h3")
			}
			title := strings.TrimSpace(titleElem.Text())

			var href string
			s.Find("a").Each(func(j int, a *goquery.Selection) {
				if href == "" {
					h, ok := a.Attr("href")
					if ok && strings.HasPrefix(h, "http") && !strings.Contains(h, "google.com") {
						href = h
					}
				}
			})

			snippet := strings.TrimSpace(s.Find(".VwiC3b").Text())
			if snippet == "" {
				snippet = strings.TrimSpace(s.Find("span").Text())
			}

			if title != "" && href != "" {
				results = append(results, models.Result{
					Title:    title,
					URL:      href,
					Content:  snippet,
					Engine:   g.Name(),
					Category: req.Category,
				})
			}
		})
		if len(results) > 0 {
			break
		}
	}

	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  results,
	}, nil
}
