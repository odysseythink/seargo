package bases

import (
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"

	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

// ResultType 表示 bases 引擎可以产出的类型化结果种类。
type ResultType string

const (
	ResultTypeDefault ResultType = ""
	ResultTypePaper   ResultType = "paper"
	ResultTypeCode    ResultType = "code"
	ResultTypeFile    ResultType = "file"
)

// ResultTypeConfig 描述如何从 JSON/XPath 原始结果中填充 typed result 的专属字段。
type ResultTypeConfig struct {
	Type ResultType

	// PaperResult 字段
	DOIQuery           string
	AuthorsQuery       string
	PDFURLQuery        string
	HTMLURLQuery       string
	JournalQuery       string
	PublisherQuery     string
	PublishedDateQuery string

	// CodeResult 字段
	RepositoryQuery   string
	CodeLanguageQuery string
	FilenameQuery     string

	// FileResult 字段（FilenameQuery 与 CodeResult 同名复用）
	FileTypeQuery  string
	FileSizeQuery  string
	MagnetURIQuery string
}

// buildTypedResult 根据 result_type 把原始数据转换为 results.Result。
func buildTypedResult(raw interface{}, cfg ResultTypeConfig, base models.Result) results.Result {
	switch cfg.Type {
	case ResultTypePaper:
		return buildPaperResult(raw, cfg, base)
	case ResultTypeCode:
		return buildCodeResult(raw, cfg, base)
	case ResultTypeFile:
		return buildFileResult(raw, cfg, base)
	default:
		return results.WrapAPIMainResult(base)
	}
}

func buildPaperResult(raw interface{}, cfg ResultTypeConfig, base models.Result) results.Result {
	r := &results.PaperResult{
		BaseResult: results.BaseResult{
			Title:    base.Title,
			URL:      base.URL,
			Content:  base.Content,
			Engine:   base.Engine,
			Category: string(base.Category),
		},
		DOI:           extractString(raw, cfg.DOIQuery),
		Journal:       extractString(raw, cfg.JournalQuery),
		Authors:       extractStrings(raw, cfg.AuthorsQuery),
		PDFURL:        extractString(raw, cfg.PDFURLQuery),
		HTMLURL:       extractString(raw, cfg.HTMLURLQuery),
		Publisher:     extractString(raw, cfg.PublisherQuery),
		PublishedDate: extractString(raw, cfg.PublishedDateQuery),
	}
	return r
}

func buildCodeResult(raw interface{}, cfg ResultTypeConfig, base models.Result) results.Result {
	r := &results.CodeResult{
		BaseResult: results.BaseResult{
			Title:    base.Title,
			URL:      base.URL,
			Content:  base.Content,
			Engine:   base.Engine,
			Category: string(base.Category),
		},
		Repository:   extractString(raw, cfg.RepositoryQuery),
		CodeLanguage: extractString(raw, cfg.CodeLanguageQuery),
		Filename:     extractString(raw, cfg.FilenameQuery),
	}
	return r
}

func buildFileResult(raw interface{}, cfg ResultTypeConfig, base models.Result) results.Result {
	r := &results.FileResult{
		BaseResult: results.BaseResult{
			Title:    base.Title,
			URL:      base.URL,
			Content:  base.Content,
			Engine:   base.Engine,
			Category: string(base.Category),
		},
		FileType:  extractString(raw, cfg.FileTypeQuery),
		Filename:  extractString(raw, cfg.FilenameQuery),
		MagnetURI: extractString(raw, cfg.MagnetURIQuery),
	}
	if s := extractString(raw, cfg.FileSizeQuery); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			r.FileSize = n
		}
	}
	return r
}

func extractString(raw interface{}, query string) string {
	if query == "" {
		return ""
	}
	switch v := raw.(type) {
	case map[string]interface{}:
		return firstString(jsonQuery(v, query))
	case *html.Node:
		return xpathGetOne(v, query)
	}
	return ""
}

func extractStrings(raw interface{}, query string) []string {
	if query == "" {
		return nil
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	vals := jsonQuery(m, query)
	var out []string
	for _, val := range vals {
		if s, ok := val.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func xpathGetOne(node *html.Node, xpath string) string {
	found, err := htmlquery.Query(node, xpath)
	if err != nil || found == nil {
		return ""
	}
	return strings.TrimSpace(htmlquery.InnerText(found))
}

func parseResultTypeConfig(extra map[string]any) ResultTypeConfig {
	if extra == nil {
		return ResultTypeConfig{}
	}
	raw, ok := extra["result_type"]
	if !ok {
		return ResultTypeConfig{}
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return ResultTypeConfig{}
	}
	get := func(key string) string {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}
	return ResultTypeConfig{
		Type:               ResultType(get("type")),
		DOIQuery:           get("doi_query"),
		AuthorsQuery:       get("authors_query"),
		PDFURLQuery:        get("pdf_url_query"),
		HTMLURLQuery:       get("html_url_query"),
		JournalQuery:       get("journal_query"),
		PublisherQuery:     get("publisher_query"),
		PublishedDateQuery: get("published_date_query"),
		RepositoryQuery:    get("repository_query"),
		CodeLanguageQuery:  get("code_language_query"),
		FilenameQuery:      get("filename_query"),
		FileTypeQuery:      get("file_type_query"),
		FileSizeQuery:      get("file_size_query"),
		MagnetURIQuery:     get("magnet_uri_query"),
	}
}
