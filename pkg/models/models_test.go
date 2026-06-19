package models

import (
    "encoding/json"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRequestCacheKey(t *testing.T) {
    // Same request → same key
    r1 := &Request{Query: "go programming", Category: CategoryGeneral, SafeSearch: 1}
    r2 := &Request{Query: "go programming", Category: CategoryGeneral, SafeSearch: 1}
    assert.Equal(t, r1.CacheKey(), r2.CacheKey())

    // Different SafeSearch → different key
    r3 := &Request{Query: "go programming", Category: CategoryGeneral, SafeSearch: 2}
    assert.NotEqual(t, r1.CacheKey(), r3.CacheKey())

    // Different query → different key
    r4 := &Request{Query: "python programming", Category: CategoryGeneral, SafeSearch: 1}
    assert.NotEqual(t, r1.CacheKey(), r4.CacheKey())

    // Different page → different key
    r5 := &Request{Query: "go programming", Category: CategoryGeneral, SafeSearch: 1, Page: 2}
    assert.NotEqual(t, r1.CacheKey(), r5.CacheKey())

    // Zero-value SafeSearch (0) should serialize correctly
    r6 := &Request{Query: "test", Category: CategoryGeneral, SafeSearch: 0}
    require.NotPanics(t, func() { r6.CacheKey() })
}

func TestCacheKeyIncludesPageSize(t *testing.T) {
    r1 := &Request{Query: "test", Category: CategoryGeneral, PageSize: 10}
    r2 := &Request{Query: "test", Category: CategoryGeneral, PageSize: 20}
    assert.NotEqual(t, r1.CacheKey(), r2.CacheKey(),
        "Different page sizes should produce different cache keys")
}

func TestAllCategories(t *testing.T) {
    cats := AllCategories()
    assert.Len(t, cats, 10)
    // First four must match original order
    assert.Equal(t, CategoryGeneral, cats[0])
    assert.Equal(t, CategoryImages, cats[1])
    assert.Equal(t, CategoryVideos, cats[2])
    assert.Equal(t, CategoryNews, cats[3])
    // Verify new categories exist
    catSet := make(map[Category]bool)
    for _, c := range cats {
        catSet[c] = true
    }
    assert.True(t, catSet[CategoryMap])
    assert.True(t, catSet[CategoryMusic])
    assert.True(t, catSet[CategoryIT])
    assert.True(t, catSet[CategoryScience])
    assert.True(t, catSet[CategoryFiles])
    assert.True(t, catSet[CategorySocialMedia])
}

func TestRequestNormalize(t *testing.T) {
	defaults := NormalizeDefaults{
		DefaultLang:     "en",
		DefaultCategory: CategoryGeneral,
		DefaultPageSize: 10,
		MaxResults:      50,
	}

	// Case 1: All zero — should fill defaults
	r1 := &Request{Query: "test"}
	r1.Normalize(defaults)
	assert.Equal(t, "en", r1.Language)
	assert.Equal(t, CategoryGeneral, r1.Category)
	assert.Equal(t, 10, r1.PageSize)
	assert.Equal(t, 1, r1.Page)

	// Case 2: User-provided values should NOT be overwritten
	r2 := &Request{
		Query: "test", Language: "zh-CN",
		Category: CategoryImages, PageSize: 20, Page: 3,
		SafeSearch: 2,
	}
	r2.Normalize(defaults)
	assert.Equal(t, "zh-CN", r2.Language)
	assert.Equal(t, CategoryImages, r2.Category)
	assert.Equal(t, 20, r2.PageSize)
	assert.Equal(t, 3, r2.Page)

	// Case 3: Page=0 should default to 1
	r3 := &Request{Query: "test", Page: 0}
	r3.Normalize(defaults)
	assert.Equal(t, 1, r3.Page)

	// Case 4: PageSize > MaxResults should be capped
	r4 := &Request{Query: "test", PageSize: 100}
	r4.Normalize(defaults)
	assert.Equal(t, 50, r4.PageSize)

	// Case 5: Negative page is clamped to 1
	r5 := &Request{Query: "test", Page: -1}
	r5.Normalize(defaults)
	assert.Equal(t, 1, r5.Page)
}

func TestCategoryValues(t *testing.T) {
    // Verify string values
    assert.Equal(t, "general", string(CategoryGeneral))
    assert.Equal(t, "images", string(CategoryImages))
    assert.Equal(t, "videos", string(CategoryVideos))
    assert.Equal(t, "news", string(CategoryNews))
    assert.Equal(t, "map", string(CategoryMap))
    assert.Equal(t, "music", string(CategoryMusic))
    assert.Equal(t, "it", string(CategoryIT))
    assert.Equal(t, "science", string(CategoryScience))
    assert.Equal(t, "files", string(CategoryFiles))
    assert.Equal(t, "social media", string(CategorySocialMedia))
}

func TestCacheKeyTableDriven(t *testing.T) {
    tests := []struct {
        name     string
        req1     Request
        req2     Request
        wantSame bool
    }{
        {
            name:     "same request → same key",
            req1:     Request{Query: "test", Category: CategoryGeneral, SafeSearch: 1, Page: 1, PageSize: 10},
            req2:     Request{Query: "test", Category: CategoryGeneral, SafeSearch: 1, Page: 1, PageSize: 10},
            wantSame: true,
        },
        {
            name:     "different query → different key",
            req1:     Request{Query: "foo", Category: CategoryGeneral},
            req2:     Request{Query: "bar", Category: CategoryGeneral},
            wantSame: false,
        },
        {
            name:     "different SafeSearch → different key",
            req1:     Request{Query: "test", SafeSearch: 0},
            req2:     Request{Query: "test", SafeSearch: 2},
            wantSame: false,
        },
        {
            name:     "different category → different key",
            req1:     Request{Query: "test", Category: CategoryGeneral},
            req2:     Request{Query: "test", Category: CategoryImages},
            wantSame: false,
        },
        {
            name:     "different page → different key",
            req1:     Request{Query: "test", Page: 1, PageSize: 10},
            req2:     Request{Query: "test", Page: 2, PageSize: 10},
            wantSame: false,
        },
        {
            name:     "different pageSize → different key",
            req1:     Request{Query: "test", Page: 1, PageSize: 10},
            req2:     Request{Query: "test", Page: 1, PageSize: 20},
            wantSame: false,
        },
        {
            name:     "zero-value SafeSearch is consistent",
            req1:     Request{Query: "test", SafeSearch: 0, PageSize: 10},
            req2:     Request{Query: "test", SafeSearch: 0, PageSize: 10},
            wantSame: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            k1 := tt.req1.CacheKey()
            k2 := tt.req2.CacheKey()
            if tt.wantSame {
                assert.Equal(t, k1, k2)
            } else {
                assert.NotEqual(t, k1, k2)
            }
        })
    }
}

func TestCategoryValidValues(t *testing.T) {
    tests := []struct {
        category Category
        valid    bool
    }{
        {CategoryGeneral, true},
        {CategoryImages, true},
        {CategoryVideos, true},
        {CategoryNews, true},
        {CategoryMap, true},
        {CategoryMusic, true},
        {CategoryIT, true},
        {CategoryScience, true},
        {CategoryFiles, true},
        {CategorySocialMedia, true},
        {Category("unknown"), false},
        {Category(""), false},
        {Category("GENERAL"), false},
    }

    validSet := make(map[Category]bool)
    for _, c := range AllCategories() {
        validSet[c] = true
    }

    for _, tt := range tests {
        t.Run(string(tt.category), func(t *testing.T) {
            assert.Equal(t, tt.valid, validSet[tt.category])
        })
    }
}

func TestResultNewFieldsJSON(t *testing.T) {
	r := Result{
		Title:   "Test",
		URL:     "https://example.com",
		Engine:  "google",
		Engines: []string{"google", "bing"},
		Score:   3.5,
		Domain:  "example.com",
	}

	data, err := json.Marshal(r)
	assert.NoError(t, err)

	var decoded Result
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "Test", decoded.Title)
	assert.Equal(t, []string{"google", "bing"}, decoded.Engines)
	assert.Equal(t, "example.com", decoded.Domain)
}

func TestResultEnginesOmitEmpty(t *testing.T) {
	r := Result{Title: "T", URL: "https://x.com"}
	data, err := json.Marshal(r)
	assert.NoError(t, err)
	assert.NotContains(t, string(data), `"engines"`)
}

func TestResponseNewFieldsJSON(t *testing.T) {
	resp := Response{
		Query:       "test",
		Results:     []Result{},
		Answers:     []Answer{{Answer: "42"}},
		Infoboxes:   []Infobox{{Title: "info", Content: "body"}},
		RedirectURL: "https://google.com/search?q=test",
		EngineData:  map[string]any{"key": "val"},
	}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var decoded Response
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Len(t, decoded.Answers, 1)
	assert.Equal(t, "42", decoded.Answers[0].Answer)
	assert.Len(t, decoded.Infoboxes, 1)
	assert.Equal(t, "https://google.com/search?q=test", decoded.RedirectURL)
}

func TestResponseNewFieldsOmitEmpty(t *testing.T) {
	resp := Response{Query: "test", Results: []Result{}}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.NotContains(t, string(data), `"answers"`)
	assert.NotContains(t, string(data), `"redirect_url"`)
}
