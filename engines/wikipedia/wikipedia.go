package wikipedia

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	engine.Register("wikipedia", &Wikipedia{})
}

// wikipediaSearchURL is the search URL template. It is a variable so tests can
// redirect requests to a local mock server. The template receives (language,
// url-encoded query) as arguments.
var wikipediaSearchURL = "https://%s.wikipedia.org/w/index.php?search=%s"

type Wikipedia struct {
	client *httpx.Client
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

func (w *Wikipedia) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (w *Wikipedia) Setup(cfg engine.EngineInitConfig) bool {
	w.client = cfg.Client
	return true
}

func (w *Wikipedia) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://en.wikipedia.org",
		WikidataID: "Q52",
	}
}

func (w *Wikipedia) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	lang := resolveLanguage(req.Language)
	searchURL := fmt.Sprintf(wikipediaSearchURL, lang, url.QueryEscape(req.Query))

	resp, err := w.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("wikipedia request failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	articleURL := resp.URL
	if articleURL == "" {
		articleURL = searchURL
	}

	var typed []results.Result

	// Wikipedia search results
	doc.Find(".mw-search-result").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find(".mw-search-result-heading a")
		title := strings.TrimSpace(titleElem.Text())
		href, _ := titleElem.Attr("href")
		if href != "" && !strings.HasPrefix(href, "http") {
			href = fmt.Sprintf("https://%s.wikipedia.org%s", lang, href)
		}
		snippet := strings.TrimSpace(s.Find(".searchresult").Text())

		if title != "" && href != "" {
			typed = append(typed, results.WrapAPIMainResult(models.Result{
				Title:    title,
				URL:      href,
				Content:  snippet,
				Engine:   w.Name(),
				Category: req.Category,
			}))
		}
	})

	// If the page itself is an article with an infobox, extract it.
	if doc.Find(".infobox").Length() > 0 {
		infobox := parseWikipediaInfobox(doc, lang, articleURL)
		typed = append(typed, &infobox)
	}

	raw := make([]any, len(typed))
	for i, r := range typed {
		raw[i] = r
	}
	return &models.Response{
		Query:        req.Query,
		Category:     req.Category,
		Results:      results.ToAPIResult(typed),
		TypedResults: raw,
	}, nil
}

func resolveLanguage(lang string) string {
	if lang == "" {
		return "en"
	}
	lang = strings.SplitN(lang, "-", 2)[0]
	lang = strings.SplitN(lang, "_", 2)[0]
	if lang == "" {
		return "en"
	}
	return lang
}

func parseWikipediaInfobox(doc *goquery.Document, lang, pageURL string) results.InfoboxResult {
	title := strings.TrimSpace(doc.Find("#firstHeading").Text())
	description := strings.TrimSpace(doc.Find(".shortdescription").Text())

	infobox := results.InfoboxResult{
		BaseResult: results.BaseResult{
			Title:    title,
			Content:  description,
			Engine:   "wikipedia",
			Template: "infobox",
			Category: string(models.CategoryGeneral),
		},
		InfoboxID: pageURL,
	}

	doc.Find(".infobox tr").Each(func(i int, s *goquery.Selection) {
		header := strings.TrimSpace(s.Find("th").Text())
		dataCell := s.Find("td")
		value := strings.TrimSpace(dataCell.Text())
		link, _ := dataCell.Find("a").Attr("href")
		if link != "" && !strings.HasPrefix(link, "http") {
			link = fmt.Sprintf("https://%s.wikipedia.org%s", lang, link)
		}

		if header != "" && value != "" {
			infobox.Attributes = append(infobox.Attributes, results.InfoboxAttribute{
				Label: header,
				Value: value,
				URL:   link,
			})
		}
	})

	if img, exists := doc.Find(".infobox img").Attr("src"); exists && img != "" {
		infobox.ImgSrc = makeAbsoluteURL(img, lang)
		infobox.ImgAlt = title
	}

	infobox.URLs = append(infobox.URLs, results.InfoboxURL{Title: "Wikipedia", URL: pageURL})

	return infobox
}

func makeAbsoluteURL(raw, lang string) string {
	if strings.HasPrefix(raw, "//") {
		return "https:" + raw
	}
	if strings.HasPrefix(raw, "/") {
		return fmt.Sprintf("https://%s.wikipedia.org%s", lang, raw)
	}
	if strings.HasPrefix(raw, "http") {
		return raw
	}
	return fmt.Sprintf("https://%s.wikipedia.org/%s", lang, raw)
}
