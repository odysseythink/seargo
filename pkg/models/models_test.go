package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestCacheKey(t *testing.T) {
	r1 := &Request{Query: "go programming", Category: CategoryGeneral}
	r2 := &Request{Query: "go programming", Category: CategoryGeneral}
	r3 := &Request{Query: "python programming", Category: CategoryGeneral}

	assert.Equal(t, r1.CacheKey(), r2.CacheKey(), "Same request should have same cache key")
	assert.NotEqual(t, r1.CacheKey(), r3.CacheKey(), "Different queries should have different cache keys")
}
