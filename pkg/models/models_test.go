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
