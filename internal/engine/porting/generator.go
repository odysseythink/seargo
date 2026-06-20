package porting

import (
	"fmt"
	"regexp"
	"strings"
)

// SkeletonResult holds the generated Go code and fixture stub for an engine.
type SkeletonResult struct {
	EngineName  string
	BaseType    string
	GoCode      string
	FixtureYAML string
}

// GenerateSkeleton analyzes a SearXNG Python engine source and produces
// a Go skeleton + golden fixture stub.
func GenerateSkeleton(engineName, pySource string) (*SkeletonResult, error) {
	baseType := detectBaseType(pySource)

	categories := extractPythonList(findPythonVar(pySource, "categories"))
	baseURL := extractPythonStringVar(pySource, "base_url")
	searchURL := extractPythonStringVar(pySource, "search_url")

	var goCode, fixtureYAML string

	switch baseType {
	case "xpath":
		goCode = generateXPathSkeleton(engineName, categories, baseURL, searchURL, pySource)
		fixtureYAML = generateXPathFixture(engineName, searchURL)
	case "json_engine":
		goCode = generateJSONSkeleton(engineName, categories, baseURL, searchURL, pySource)
		fixtureYAML = generateJSONFixture(engineName, searchURL)
	case "mediawiki":
		goCode = generateMediaWikiSkeleton(engineName, categories, baseURL)
		fixtureYAML = generateMediaWikiFixture(engineName, baseURL)
	default:
		goCode = generateCustomSkeleton(engineName, pySource)
		fixtureYAML = generateCustomFixture(engineName)
	}

	return &SkeletonResult{
		EngineName:  engineName,
		BaseType:    baseType,
		GoCode:      goCode,
		FixtureYAML: fixtureYAML,
	}, nil
}

// detectBaseType determines which base engine to use based on Python source patterns.
func detectBaseType(pySource string) string {
	if strings.Contains(pySource, "results_xpath") || strings.Contains(pySource, "url_xpath") {
		return "xpath"
	}
	if strings.Contains(pySource, "results_query") || strings.Contains(pySource, "url_query") {
		return "json_engine"
	}
	if strings.Contains(pySource, "action=query") || strings.Contains(pySource, "list=search") || strings.Contains(pySource, "w/api.php") {
		return "mediawiki"
	}
	return "custom"
}

// ---- Regex-based extraction helpers ----

func findPythonVar(pySource, varName string) string {
	re := regexp.MustCompile(fmt.Sprintf(`%s\s*=\s*\[([^\]]*)\]`, varName))
	m := re.FindStringSubmatch(pySource)
	if len(m) > 1 {
		return "[" + m[1] + "]"
	}
	return ""
}

func extractPythonStringVar(pySource, varName string) string {
	// Match: var = "value"
	re := regexp.MustCompile(fmt.Sprintf(`%s\s*=\s*(?:"([^"]*)"|'([^']*)')`, varName))
	m := re.FindStringSubmatch(pySource)
	if len(m) > 1 {
		if m[1] != "" {
			return m[1]
		}
		if len(m) > 2 && m[2] != "" {
			return m[2]
		}
	}
	// Match: var = base_url + "/path"
	re = regexp.MustCompile(fmt.Sprintf(`%s\s*=\s*base_url\s*\+\s*"([^"]*)"`, varName))
	m = re.FindStringSubmatch(pySource)
	if len(m) > 1 {
		baseURL := extractPythonStringVar(pySource, "base_url")
		if baseURL != "" {
			return strings.TrimSuffix(baseURL, "/") + m[1]
		}
	}
	return ""
}

func extractPythonList(listStr string) []string {
	if !strings.HasPrefix(listStr, "[") {
		return nil
	}
	inner := strings.TrimPrefix(listStr, "[")
	inner = strings.TrimSuffix(inner, "]")
	if inner == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"'`)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ---- Skeleton generators ----

func generateXPathSkeleton(name string, categories []string, baseURL, searchURL string, pySource string) string {
	if searchURL == "" && baseURL != "" {
		searchURL = baseURL + "/search?q={query}"
	}

	resultXPath := extractPythonStringVar(pySource, "results_xpath")
	urlXPath := extractPythonStringVar(pySource, "url_xpath")
	titleXPath := extractPythonStringVar(pySource, "title_xpath")
	contentXPath := extractPythonStringVar(pySource, "content_xpath")

	return fmt.Sprintf(`package %s

import (
	"github.com/seargo/seargo/internal/engine/bases"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine := bases.NewXPathEngine("%s", []models.Category{%s}, bases.XPathConfig{
		SearchURL:    %q,
		ResultXPath:  %q,
		URLXPath:     %q,
		TitleXPath:   %q,
		ContentXPath: %q,
	})
	_ = engine
}
`, name, name, formatCategories(categories), searchURL, resultXPath, urlXPath, titleXPath, contentXPath)
}

func generateJSONSkeleton(name string, categories []string, baseURL, searchURL string, pySource string) string {
	if searchURL == "" && baseURL != "" {
		searchURL = baseURL + "?q={query}"
	}

	resultsQuery := extractPythonStringVar(pySource, "results_query")
	urlQuery := extractPythonStringVar(pySource, "url_query")
	titleQuery := extractPythonStringVar(pySource, "title_query")
	contentQuery := extractPythonStringVar(pySource, "content_query")

	return fmt.Sprintf(`package %s

import (
	"github.com/seargo/seargo/internal/engine/bases"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine := bases.NewJSONEngine("%s", []models.Category{%s}, bases.JSONEngineConfig{
		SearchURL:    %q,
		ResultsQuery: %q,
		URLQuery:     %q,
		TitleQuery:   %q,
		ContentQuery: %q,
	})
	_ = engine
}
`, name, name, formatCategories(categories), searchURL, resultsQuery, urlQuery, titleQuery, contentQuery)
}

func generateMediaWikiSkeleton(name string, categories []string, baseURL string) string {
	if baseURL == "" {
		baseURL = "https://en.wikipedia.org/w/api.php"
	}

	return fmt.Sprintf(`package %s

import (
	"github.com/seargo/seargo/internal/engine/bases"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine := bases.NewMediaWikiEngine("%s", []models.Category{%s}, bases.MediaWikiConfig{
		BaseURL: %q,
	})
	_ = engine
}
`, name, name, formatCategories(categories), baseURL)
}

func generateCustomSkeleton(name string, pySource string) string {
	srcTruncated := pySource
	if len(srcTruncated) > 500 {
		srcTruncated = srcTruncated[:500]
	}

	camelName := toCamel(name)

	return fmt.Sprintf(`package %s

import (
	"context"
	"fmt"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

// Manual porting required for engine %q — custom engine skeleton.
// Original Python source excerpt:
// %s

type %sEngine struct {
	client *httpx.Client
}

func init() {
	engine.Register("%s", &%sEngine{})
}

func (e *%sEngine) Name() string                     { return "%s" }
func (e *%sEngine) Categories() []models.Category     { return nil }
func (e *%sEngine) Capabilities() engine.Capabilities { return engine.Capabilities{} }
func (e *%sEngine) About() engine.EngineAbout         { return engine.EngineAbout{} }
func (e *%sEngine) Setup(cfg engine.EngineInitConfig) bool { return true }
func (e *%sEngine) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }
func (e *%sEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return nil, fmt.Errorf("engine %%s: manual porting required", "%s")
}
`,
		name,          // 1: package name (package %s)
		name,          // 2: engine name (%q comment)
		srcTruncated,  // 3: truncated source (// %s)
		camelName,     // 4: type Name (%sEngine)
		name,          // 5: Register("%s"
		camelName,     // 6: &%sEngine{}
		camelName,     // 7: (e *%sEngine) Name
		name,          // 8: return "%s"
		camelName,     // 9: (e *%sEngine) Categories
		camelName,     // 10: (e *%sEngine) Capabilities
		camelName,     // 11: (e *%sEngine) About
		camelName,     // 12: (e *%sEngine) Setup
		camelName,     // 13: (e *%sEngine) Init
		camelName,     // 14: (e *%sEngine) Search
		name,          // 15: "%s" in Errorf
	)
}

// ---- Fixture generators ----

func generateXPathFixture(name, searchURL string) string {
	return fmt.Sprintf(`# Golden fixture for %s (xpath base)
engine: %s
request:
  query: "test query"
  category: general
mock_response:
  status: 200
  headers:
    Content-Type: text/html
  body: |
    <html><body>
      <div class="result">
        <h3><a href="https://example.com/1">Result 1</a></h3>
        <p class="snippet">Snippet one</p>
      </div>
    </body></html>
expected_results:
  - title: "Result 1"
    url: "https://example.com/1"
    content: "Snippet one"
`, name, name)
}

func generateJSONFixture(name, searchURL string) string {
	return fmt.Sprintf(`# Golden fixture for %s (json_engine base)
engine: %s
request:
  query: "test"
  category: general
mock_response:
  status: 200
  headers:
    Content-Type: application/json
  body: |
    {"response":{"docs":[{"title":"R1","url":"https://x.com/1","snippet":"S1"}]}}
expected_results:
  - title: "R1"
    url: "https://x.com/1"
    content: "S1"
`, name, name)
}

func generateMediaWikiFixture(name, baseURL string) string {
	return fmt.Sprintf(`# Golden fixture for %s (mediawiki base)
engine: %s
request:
  query: "test"
  category: general
mock_response:
  status: 200
  headers:
    Content-Type: application/json
  body: |
    {"query":{"search":[{"title":"Test","pageid":1,"snippet":"Snippet"}]}}
expected_results:
  - title: "Test"
`, name, name)
}

func generateCustomFixture(name string) string {
	return fmt.Sprintf(`# Golden fixture for %s (custom engine)
# TODO: fill in mock response and expected results after manual porting
engine: %s
request:
  query: "test"
  category: general
mock_response: {}
expected_results: []
`, name, name)
}

// ---- Helpers ----

func formatCategories(cats []string) string {
	if len(cats) == 0 {
		return ""
	}
	quoted := make([]string, len(cats))
	for i, c := range cats {
		quoted[i] = fmt.Sprintf("models.Category%s", toCamel(c))
	}
	return strings.Join(quoted, ", ")
}

func toCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}
