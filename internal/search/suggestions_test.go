package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeSuggestions(t *testing.T) {
	result := mergeSuggestions([][]string{
		{"hello", "world"},
		{"hello", "foo"},
		nil,
		{"bar", "baz"},
	})
	assert.Equal(t, []string{"hello", "world", "foo", "bar", "baz"}, result)
}

func TestMergeSuggestionsDedupCaseInsensitive(t *testing.T) {
	result := mergeSuggestions([][]string{
		{"Hello", "World"},
		{"hello", "WORLD"},
	})
	assert.Equal(t, []string{"Hello", "World"}, result)
}

func TestMergeSuggestionsEmpty(t *testing.T) {
	result := mergeSuggestions([][]string{})
	assert.Nil(t, result)

	result = mergeSuggestions([][]string{nil, nil})
	assert.Nil(t, result)
}

func TestMergeSuggestionsLimit(t *testing.T) {
	input := make([][]string, 1)
	input[0] = make([]string, 15)
	for i := 0; i < 15; i++ {
		input[0][i] = string(rune('a' + i))
	}
	result := mergeSuggestions(input)
	assert.Len(t, result, 10, "should be limited to 10")
}
