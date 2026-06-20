package porting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateXPathEngine(t *testing.T) {
	pySource := `
base_url = "https://example.com/search"
search_url = base_url + "?q={query}"
results_xpath = "//div[@class='result']"
url_xpath = ".//a/@href"
title_xpath = ".//h3/a"
content_xpath = ".//p[@class='snippet']"
categories = ["general"]
paging = True
`

	result, err := GenerateSkeleton("test_engine", pySource)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.BaseType, "xpath")
	assert.Contains(t, result.GoCode, "NewXPathEngine")
	assert.Contains(t, result.GoCode, `SearchURL:`)
	assert.Contains(t, result.GoCode, `ResultXPath:`)

	assert.NotEmpty(t, result.FixtureYAML)
	assert.Contains(t, result.FixtureYAML, "test_engine")
}

func TestGenerator_GenerateJSONEngine(t *testing.T) {
	pySource := `
base_url = "https://api.example.com"
search_url = base_url + "/search?q={query}"
results_query = "response/docs"
url_query = "url"
title_query = "title"
content_query = "snippet"
categories = ["general"]
`

	result, err := GenerateSkeleton("json_engine_test", pySource)
	require.NoError(t, err)

	assert.Contains(t, result.BaseType, "json_engine")
	assert.Contains(t, result.GoCode, "NewJSONEngine")
	assert.Contains(t, result.GoCode, `ResultsQuery:`)
}

func TestGenerator_GenerateMediaWikiEngine(t *testing.T) {
	pySource := `
base_url = "https://en.wikipedia.org/w/api.php"
categories = ["general"]
`

	result, err := GenerateSkeleton("wiki_test", pySource)
	require.NoError(t, err)

	assert.Contains(t, result.BaseType, "mediawiki")
	assert.Contains(t, result.GoCode, "NewMediaWikiEngine")
}

func TestGenerator_UnknownBase_FallbackToCustom(t *testing.T) {
	pySource := `
def request(query, params):
    return params

def response(resp):
    return []
`

	result, err := GenerateSkeleton("custom_eng", pySource)
	require.NoError(t, err)

	assert.Contains(t, result.BaseType, "custom")
	assert.Contains(t, result.GoCode, "custom engine skeleton")
}

func TestGenerator_ExtractCategories(t *testing.T) {
	cats := extractPythonList(`['general', 'images', 'news']`)
	assert.Equal(t, []string{"general", "images", "news"}, cats)
}

func TestGenerator_ExtractStringVar(t *testing.T) {
	val := extractPythonStringVar(`search_url = "https://example.com/search?q={query}"`, "search_url")
	assert.Equal(t, "https://example.com/search?q={query}", val)
}
