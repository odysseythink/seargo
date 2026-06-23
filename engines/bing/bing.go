package bing

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"unicode"

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
		SupportsTimeRange:  false,
	}
}

func (b *Bing) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

func (b *Bing) Setup(cfg engine.EngineInitConfig) bool {
	b.client = cfg.Client
	return true
}

func (b *Bing) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://www.bing.com",
		WikidataID: "Q182496",
	}
}

var safeSearchMap = map[int]string{
	0: "off",
	1: "moderate",
	2: "strict",
}

// bingSafeSearch clamps req.SafeSearch to valid values [0,2].
func bingSafeSearch(level int) string {
	if level < 0 || level > 2 {
		level = 1 // moderate
	}
	return safeSearchMap[level]
}

func (b *Bing) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	queryParams := url.Values{}
	queryParams.Set("q", req.Query)
	queryParams.Set("adlt", bingSafeSearch(req.SafeSearch))

	mkt := resolveBingMKT(req)
	if mkt != "" {
		queryParams.Set("mkt", mkt)
	}

	searchURL := "https://cn.bing.com/search?" + queryParams.Encode()

	rb := b.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8").
		SetHeader("Accept-Language", bingAcceptLanguage(mkt)).
		SetHeader("DNT", "1").
		SetHeader("Cache-Control", "max-age=0")

	resp, err := rb.Get(searchURL)
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
		title := normalizeBingText(titleElem.Text())
		href, _ := titleElem.Attr("href")
		href = decodeBingRedirect(href)

		// Bing usually places the snippet in the first <p>; later paragraphs are
		// entity cards / related links. Use the first non-empty paragraph to match
		// SearXNG's extraction.
		s.Find(".algoSlug_icon").Remove()
		snippet := ""
		s.Find("p").Each(func(_ int, p *goquery.Selection) {
			if snippet != "" {
				return
			}
			text := normalizeBingText(p.Text())
			if text != "" {
				snippet = text
			}
		})
		// SearXNG truncates Bing snippets at roughly 200 UTF-8 bytes.
		snippet = truncateSnippetBytes(snippet, 200)

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

func resolveBingMKT(req *models.Request) string {
	if req.Locale != "" && req.Locale != "clear" {
		return req.Locale
	}
	if req.Language != "" && req.Language != "clear" {
		return req.Language
	}
	return ""
}

func bingAcceptLanguage(mkt string) string {
	if mkt == "" || mkt == "clear" {
		return "en-US,en;q=0.9"
	}
	lang := mkt
	if idx := strings.Index(mkt, "-"); idx > 0 {
		lang = mkt[:idx]
	}
	return fmt.Sprintf("%s,%s;q=0.9", mkt, lang)
}

func decodeBingRedirect(href string) string {
	if !strings.HasPrefix(href, "https://cn.bing.com/ck/a?") {
		return href
	}
	u, err := url.Parse(href)
	if err != nil {
		return href
	}
	uVal := u.Query().Get("u")
	if uVal == "" || !strings.HasPrefix(uVal, "a1") {
		return href
	}
	encoded := uVal[2:]
	encoded += strings.Repeat("=", (-len(encoded))%4)
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return href
	}
	return string(decoded)
}

// normalizeBingText collapses all Unicode whitespace variants to a single ASCII
// space and trims, matching SearXNG's extract_text output.
func normalizeBingText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) {
			b.WriteByte(' ')
		} else {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// truncateSnippet truncates text to maxRunes characters at a word boundary,
// appending an ellipsis when truncated.
func truncateSnippet(text string, maxRunes int) string {
	if len([]rune(text)) <= maxRunes {
		return text
	}
	runes := []rune(text)
	// find last space before limit
	cut := maxRunes
	for i := maxRunes; i >= 0; i-- {
		if runes[i] == ' ' {
			cut = i
			break
		}
	}
	if cut <= 0 {
		cut = maxRunes
	}
	return string(runes[:cut]) + "…"
}

// truncateSnippetBytes truncates text to maxBytes UTF-8 bytes at a word
// boundary, appending an ellipsis when truncated. This matches SearXNG's
// snippet length for CJK and Latin text.
func truncateSnippetBytes(text string, maxBytes int) string {
	if len(text) <= maxBytes {
		return text
	}
	b := []byte(text)
	cut := maxBytes
	// walk back to a space boundary without breaking a multi-byte rune
	for cut > 0 {
		if b[cut] == ' ' {
			break
		}
		// don't slice inside a UTF-8 continuation byte
		if b[cut]&0xC0 == 0x80 {
			cut--
			continue
		}
		cut--
	}
	if cut <= 0 {
		cut = maxBytes
	}
	return string(b[:cut]) + "…"
}

