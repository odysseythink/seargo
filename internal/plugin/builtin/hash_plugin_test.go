package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/stretchr/testify/assert"
)

func TestHashPlugin_MD5(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "md5 hello"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, `md5("hello") = 5d41402abc4b2a76b9719d911017c592`, results[0].Title)
	assert.Equal(t, "5d41402abc4b2a76b9719d911017c592", results[0].Content)
	assert.Equal(t, "hash_plugin", results[0].Engine)
}

func TestHashPlugin_SHA256(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "sha256 hello"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, `sha256("hello") = 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824`, results[0].Title)
}

func TestHashPlugin_SHA1(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "sha1 hello"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, `sha1("hello") = aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d`, results[0].Title)
}

func TestHashPlugin_MD5MultiWord(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "md5 hello world"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, `md5("hello world")`)
	assert.Equal(t, "5eb63bbbe01eeed093cb22bb8f5acdc3", results[0].Content)
}

func TestHashPlugin_NoMatch(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "regular search"}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestHashPlugin_UnknownAlgorithm(t *testing.T) {
	p := &hashPlugin{}
	// The regex won't match unknown algorithms since it's hardcoded
	ctx := &plugin.SearchContext{Query: "unknown hello"}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestHashPlugin_CaseInsensitive(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "MD5 Hello"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, `md5("Hello") = 8b1a9953c4611296a827abf8c47804d7`, results[0].Title)
}

func TestHashPlugin_SHA224(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "sha224 hello"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, `sha224("hello")`)
}

func TestHashPlugin_SHA384(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "sha384 hello"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, `sha384("hello")`)
}

func TestHashPlugin_SHA512(t *testing.T) {
	p := &hashPlugin{}
	ctx := &plugin.SearchContext{Query: "sha512 hello"}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, `sha512("hello")`)
}
