package models

import (
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
