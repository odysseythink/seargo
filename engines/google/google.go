package google

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/search/processor"
	"github.com/seargo/seargo/pkg/models"
	"github.com/odysseythink/mlog"
)

func init() {
	engine.Register("google", &Google{})
}

// CaptchaError signals that Google returned a bot-protection / sorry page.
type CaptchaError struct{ Msg string }

func (e *CaptchaError) Error() string {
	return "google captcha: " + e.Msg
}

type googleInfo struct {
	language  string
	country   string
	subdomain string
	params    map[string]string
	headers   map[string]string
	cookies   map[string]string
}

type Google struct {
	client *httpx.Client
	cfg    config.GoogleEngineParams
	traits engine.EngineTraits
	uaPool *httpx.UserAgentPool
}

var (
	timeRangeMapping = map[string]string{
		"day":   "d",
		"week":  "w",
		"month": "m",
		"year":  "y",
	}
	safeSearchMapping = map[int]string{
		0: "off",
		1: "medium",
		2: "high",
	}
	dataImageRE = regexp.MustCompile(`(data:image[^']*?)'[^']*?'((?:dimg|pimg|tsuid)[^']*)`)
)

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
	g.cfg = cfg.GoogleParams
	g.traits = cfg.EngineTraits
	pool, err := httpx.NewUserAgentPool("data/useragents.json")
	if err != nil {
		mlog.Warning("failed to load useragents.json for google engine, using fallback", "error", err)
	}
	g.uaPool = pool
	return true
}

func (g *Google) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://www.google.com",
		WikidataID: "Q95",
	}
}

// safeSearchParam clamps safeSearch to valid values [0,2] and returns
// the corresponding Google safe parameter value.
func safeSearchParam(safeSearch int) string {
	if safeSearch < 0 || safeSearch > 2 {
		safeSearch = 1 // default to moderate
	}
	return safeSearchMapping[safeSearch]
}

func (g *Google) buildSearchURL(query string, page int, timeRange string, safeSearch int, info googleInfo) string {
	start := (page - 1) * 10
	if start < 0 {
		start = 0
	}

	v := url.Values{}
	v.Set("q", query)
	for key, val := range info.params {
		if val != "" {
			v.Set(key, val)
		}
	}
	v.Set("filter", "0")
	v.Set("start", strconv.Itoa(start))
	if tr, ok := timeRangeMapping[timeRange]; ok {
		v.Set("tbs", "qdr:"+tr)
	}
	if ss := safeSearchParam(safeSearch); ss != "" {
		v.Set("safe", ss)
	}

	u := "https://" + info.subdomain + "/search?" + v.Encode()
	return u
}

func (g *Google) candidateURLs(baseURL, query string, cfg config.GoogleEngineParams) []string {
	urls := []string{baseURL + "&udm=14"}
	urls = append(urls, "https://www.google.com/search?hl=en&gws_rd=ssl&q="+url.QueryEscape(query))
	if cfg.UseMobileUI {
		urls = append(urls, "https://www.google.com/search?hl=en&gbv=1&q="+url.QueryEscape(query))
	}
	urls = append(urls, baseURL+"&gbv=1")
	return urls
}

func (g *Google) tryEndpoints(ctx context.Context, baseURL string, info googleInfo, query string, cfg config.GoogleEngineParams) (*httpx.Response, error) {
	var lastResp *httpx.Response
	var lastErr error
	for _, u := range g.candidateURLs(baseURL, query, cfg) {
		resp, err := g.doRequest(ctx, u, info)
		lastResp = resp
		lastErr = err
		if err == nil && !g.detectSorry(resp) {
			return resp, nil
		}
	}
	return lastResp, lastErr
}

func (g *Google) doRequest(ctx context.Context, urlStr string, info googleInfo) (*httpx.Response, error) {
	consent := info.cookies["CONSENT"]
	ua := httpx.GenGSAUserAgent(g.uaPool)
	req := g.client.R().SetContext(ctx).
		SetHeader("Accept", info.headers["Accept"]).
		SetHeader("User-Agent", ua)
	cookies := []string{"CONSENT=" + consent}
	if sgss, ok := info.cookies["SG_SS"]; ok && sgss != "" {
		cookies = append(cookies, "SG_SS="+sgss)
	}
	req.SetHeader("Cookie", strings.Join(cookies, "; "))
	return req.Get(urlStr)
}

func (g *Google) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	userLocale := req.Language
	if v, ok := ctx.Value(processor.CtxKeyUserLocale).(string); ok && v != "" {
		userLocale = v
	}
	if userLocale == "" {
		userLocale = "all"
	}

	info := g.googleInfo(userLocale, g.traits, g.cfg)
	searchURL := g.buildSearchURL(req.Query, req.Page, req.TimeRange, req.SafeSearch, info)

	resp, err := g.tryEndpoints(ctx, searchURL, info, req.Query, g.cfg)
	if err != nil {
		return nil, fmt.Errorf("google request failed: %w", err)
	}

	if g.detectSorry(resp) {
		return nil, &CaptchaError{Msg: "detected sorry page"}
	}

	results, suggestions := g.parseResults(resp)
	for i := range results {
		results[i].Category = req.Category
	}

	return &models.Response{
		Query:       req.Query,
		Category:    req.Category,
		Results:     results,
		Suggestions: suggestions,
	}, nil
}

// detectSorry checks whether the response indicates a Google CAPTCHA / sorry page.
func (g *Google) detectSorry(resp *httpx.Response) bool {
	if resp == nil {
		return false
	}
	if u, err := url.Parse(resp.URL); err == nil {
		if u.Host == "sorry.google.com" || strings.HasPrefix(u.Path, "/sorry") {
			return true
		}
	}
	if len(resp.Body) < 2000 && strings.Contains(string(resp.Body), "/sorry/") {
		return true
	}
	return false
}

// parseDataImages extracts data:image references from inline JavaScript assignments.
func (g *Google) parseDataImages(body string) map[string]string {
	m := make(map[string]string)
	for _, match := range dataImageRE.FindAllStringSubmatch(body, -1) {
		raw := match[1]
		id := match[2]
		if decoded, err := strconv.Unquote(`"` + raw + `"`); err == nil {
			m[id] = decoded
		} else {
			m[id] = raw
		}
	}
	return m
}

// parseResults extracts search results and suggestions from Google's HTML response.
func (g *Google) parseResults(resp *httpx.Response) ([]models.Result, []string) {
	body := resp.String()
	dataMap := g.parseDataImages(body)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, nil
	}

	var results []models.Result

	// Modern result container: .g with a link and snippet.
	doc.Find("div.g").Each(func(i int, gSel *goquery.Selection) {
		link := gSel.Find("a[href]").FilterFunction(func(_ int, s *goquery.Selection) bool {
			href, _ := s.Attr("href")
			return strings.HasPrefix(href, "/url?q=") || strings.HasPrefix(href, "http")
		}).First()
		if link.Length() == 0 {
			return
		}

		title := strings.TrimSpace(link.Find("h3").Text())
		if title == "" {
			title = strings.TrimSpace(link.Text())
		}
		if title == "" {
			return
		}

		rawURL, _ := link.Attr("href")
		resultURL := g.resolveResultURL(rawURL)
		if resultURL == "" {
			return
		}

		content := strings.TrimSpace(gSel.Find("div.VwiC3b").Text())
		if content == "" {
			content = strings.TrimSpace(gSel.Find("div.s3v94d, div.ILfuVd").Text())
		}

		thumbnail := ""
		if img := link.Find("img").First(); img.Length() > 0 {
			thumbnail, _ = img.Attr("src")
		}
		if thumbnail == "" {
			if img := gSel.Find("img").First(); img.Length() > 0 {
				thumbnail, _ = img.Attr("src")
			}
		}
		if strings.HasPrefix(thumbnail, "data:image") {
			if id, exists := link.Find("img").First().Attr("id"); exists {
				if real, ok := dataMap[id]; ok {
					thumbnail = real
				}
			}
		}

		results = append(results, models.Result{
			Title:        title,
			URL:          resultURL,
			Content:      content,
			Engine:       g.Name(),
			ThumbnailURL: thumbnail,
		})
	})

	// Legacy fallback: direct <a data-ved> scan.
	if len(results) == 0 {
		doc.Find("a[data-ved]").Each(func(i int, s *goquery.Selection) {
			if class, exists := s.Attr("class"); exists && class != "" {
				return
			}
			titleDiv := s.Find("div[style]").First()
			if titleDiv.Length() == 0 {
				return
			}
			title := strings.TrimSpace(titleDiv.Text())

			rawURL, exists := s.Attr("href")
			if !exists {
				return
			}
			resultURL := g.resolveResultURL(rawURL)
			if resultURL == "" {
				return
			}

			content := ""
			s.Parent().Find("div.ilUpNd.H66NU.aSRlid").Each(func(_ int, c *goquery.Selection) {
				c.Find("script").Remove()
				if content == "" {
					content = strings.TrimSpace(c.Text())
				}
			})

			thumbnail := ""
			if img := s.Find("img").First(); img.Length() > 0 {
				if src, exists := img.Attr("src"); exists {
					thumbnail = src
					if strings.HasPrefix(thumbnail, "data:image") {
						if id, exists := img.Attr("id"); exists {
							if real, ok := dataMap[id]; ok {
								thumbnail = real
							}
						}
					}
				}
			}

			results = append(results, models.Result{
				Title:        title,
				URL:          resultURL,
				Content:      content,
				Engine:       g.Name(),
				ThumbnailURL: thumbnail,
			})
		})
	}

	var suggestions []string
	doc.Find("div.gGQDvd.iIWm4b a").Each(func(i int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			suggestions = append(suggestions, text)
		}
	})

	return results, suggestions
}

func (g *Google) resolveResultURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "/url?q=") {
		q := strings.TrimPrefix(rawURL, "/url?q=")
		if idx := strings.Index(q, "&"); idx != -1 {
			q = q[:idx]
		}
		if u, err := url.QueryUnescape(q); err == nil {
			return u
		}
		return q
	}
	if strings.HasPrefix(rawURL, "http") {
		return rawURL
	}
	return ""
}

func (g *Google) googleInfo(userLocale string, traits engine.EngineTraits, cfg config.GoogleEngineParams) googleInfo {
	info := googleInfo{
		params:  make(map[string]string),
		headers: make(map[string]string),
		cookies: make(map[string]string),
	}

	resolved := traits.Resolve(userLocale)
	if resolved.All || resolved.Language == "" {
		info.language = "lang_en"
		info.country = traits.AllLocale
	} else {
		info.language = resolved.Language
		info.country = resolved.Region
	}
	if info.country == "" {
		info.country = "US"
	}

	langCode := strings.TrimPrefix(info.language, "lang_")
	if langCode == "" {
		langCode = "en"
	}

	info.subdomain = "www.google.com"
	if raw, ok := traits.Custom["supported_domains"]; ok {
		if domains, ok := raw.(map[string]any); ok {
			if v, ok := domains[info.country]; ok {
				if s, ok := v.(string); ok {
					info.subdomain = s
				}
			}
		}
	}

	info.params["hl"] = langCode + "-" + info.country
	info.params["lr"] = info.language
	if userLocale == "all" {
		info.params["lr"] = ""
	}
	info.params["cr"] = ""
	if strings.Contains(userLocale, "-") {
		info.params["cr"] = "country" + info.country
	}
	info.params["ie"] = "utf8"
	info.params["oe"] = "utf8"
		// Merge extra params from config; user-provided values override the defaults.
		for _, ep := range cfg.ExtraParams {
		if k, v, ok := strings.Cut(ep, "="); ok && k != "" {
			info.params[k] = v
		}
		}

	info.headers["Accept"] = "*/*"

	consent := cfg.ConsentCookie
	if consent == "" {
		consent = "YES+"
	}
	info.cookies["CONSENT"] = consent
	if sgss := cfg.SGSSCookie; sgss != "" {
		info.cookies["SG_SS"] = sgss
	}

	return info
}
