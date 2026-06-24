package configured

import (
	"context"
	"strconv"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/engine/bases"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	// 注册 P2 中可完全由 bases JSON/XPath 驱动的引擎。
	register("hoogle", []models.Category{models.CategoryIT})
	register("mdn", []models.Category{models.CategoryIT})
	register("mankier", []models.Category{models.CategoryIT})
	register("openairedatasets", []models.Category{models.CategoryScience})
	register("openairepublications", []models.Category{models.CategoryScience})
	// stackexchange 引擎的多实例别名：Loader 根据配置文件中的 engine: stackexchange
	// 字段和 api_site extra 参数区分实例，此处仅注册名称以通过引擎注册表检查。
	engine.Register("stackoverflow", &configuredEngine{
		name:              "stackoverflow",
		defaultCategories: []models.Category{models.CategoryIT},
	})
	engine.Register("askubuntu", &configuredEngine{
		name:              "askubuntu",
		defaultCategories: []models.Category{models.CategoryIT},
	})
	engine.Register("superuser", &configuredEngine{
		name:              "superuser",
		defaultCategories: []models.Category{models.CategoryIT},
	})
	engine.Register("wikicommons_files", &configuredEngine{
		name:              "wikicommons_files",
		defaultCategories: []models.Category{models.CategoryFiles},
	})
}

func register(name string, cats []models.Category) {
	engine.Register(name, &configuredEngine{
		name:              name,
		defaultCategories: cats,
	})
}

// configuredEngine wraps a bases JSON/XPath engine and applies SearXNG-style
// post-processing (url_prefix, content_html_to_text, {pageno} placeholders).
type configuredEngine struct {
	name              string
	defaultCategories []models.Category
	categories        []models.Category
	client            *httpx.Client
	inner             engine.Engine
	urlPrefix         string
	contentHTMLToText bool
}

func (e *configuredEngine) Name() string { return e.name }

func (e *configuredEngine) Categories() []models.Category {
	if len(e.categories) > 0 {
		return e.categories
	}
	return e.defaultCategories
}

func (e *configuredEngine) About() engine.EngineAbout { return engine.EngineAbout{} }

func (e *configuredEngine) Capabilities() engine.Capabilities {
	if e.inner != nil {
		return e.inner.Capabilities()
	}
	return engine.Capabilities{}
}

func (e *configuredEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }

func (e *configuredEngine) Setup(cfg engine.EngineInitConfig) bool {
	e.client = cfg.Client
	if len(cfg.Categories) > 0 {
		e.categories = cfg.Categories
	}

	extra := cfg.Extra
	if extra == nil {
		extra = map[string]any{}
	}

	e.urlPrefix = getString(extra, "url_prefix")
	e.contentHTMLToText = getBool(extra, "content_html_to_text")

	searchURL := normalizePageno(getString(extra, "search_url"))
	if searchURL == "" {
		return false
	}

	innerCfg := engine.EngineInitConfig{
		Client: cfg.Client,
		Extra:  cfg.Extra,
	}

	if xpath := getString(extra, "result_xpath"); xpath != "" {
		e.inner = bases.NewXPathEngine(e.name, e.Categories(), bases.XPathConfig{
			SearchURL:      searchURL,
			ResultXPath:    xpath,
			URLXPath:       getString(extra, "url_xpath"),
			TitleXPath:     getString(extra, "title_xpath"),
			ContentXPath:   getString(extra, "content_xpath"),
			ThumbnailXPath: getString(extra, "thumbnail_xpath"),
			Paging:         getBool(extra, "paging"),
			PageSize:       getInt(extra, "page_size", 10),
			FirstPage:      getInt(extra, "first_page", 1),
		})
	} else if resultsQuery := getString(extra, "results_query"); resultsQuery != "" {
		e.inner = bases.NewJSONEngine(e.name, e.Categories(), bases.JSONEngineConfig{
			SearchURL:    searchURL,
			ResultsQuery: resultsQuery,
			URLQuery:     getString(extra, "url_query"),
			TitleQuery:   getString(extra, "title_query"),
			ContentQuery: getString(extra, "content_query"),
			Paging:       getBool(extra, "paging"),
			PageSize:     getInt(extra, "page_size", 10),
			PageField:    "page",
		})
	} else {
		return false
	}

	return e.inner.Setup(innerCfg)
}

func (e *configuredEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	if e.inner == nil {
		return &models.Response{Query: req.Query, Category: req.Category}, nil
	}
	resp, err := e.inner.Search(ctx, req)
	if err != nil {
		return nil, err
	}
	e.postProcess(resp)
	return resp, nil
}

func (e *configuredEngine) postProcess(resp *models.Response) {
	if resp == nil {
		return
	}
	if e.urlPrefix == "" && !e.contentHTMLToText {
		return
	}

	for i := range resp.Results {
		if e.urlPrefix != "" && !strings.HasPrefix(resp.Results[i].URL, "http") {
			resp.Results[i].URL = bases.ResolveURL(e.urlPrefix, resp.Results[i].URL)
		}
		if e.contentHTMLToText {
			resp.Results[i].Content = bases.HTMLToText(resp.Results[i].Content)
		}
	}

	for _, raw := range resp.TypedResults {
		r, ok := raw.(results.Result)
		if !ok {
			continue
		}
		br := r.Base()
		if br == nil {
			continue
		}
		if e.urlPrefix != "" && !strings.HasPrefix(br.URL, "http") {
			br.URL = bases.ResolveURL(e.urlPrefix, br.URL)
		}
		if e.contentHTMLToText {
			br.Content = bases.HTMLToText(br.Content)
		}
	}
}

func normalizePageno(s string) string {
	return strings.ReplaceAll(s, "{pageno}", "{page}")
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getInt(m map[string]any, key string, def int) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		case string:
			if i, err := strconv.Atoi(n); err == nil {
				return i
			}
		}
	}
	return def
}
