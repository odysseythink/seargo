package bases

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustUnmarshal(t *testing.T, raw string) interface{} {
	t.Helper()
	var v interface{}
	err := json.Unmarshal([]byte(raw), &v)
	require.NoError(t, err)
	return v
}

func TestJSONQuery_DocumentsTitle(t *testing.T) {
	// SearXNG design verified case: {"documents":[{"title":"A"},{"title":"B"}]} / "documents/title" → ["A","B"]
	data := mustUnmarshal(t, `{"documents":[{"title":"A"},{"title":"B"}]}`)
	results := jsonQuery(data, "documents/title")
	assert.Equal(t, []interface{}{"A", "B"}, results)
}

func TestJSONQuery_ArrayOfObjects(t *testing.T) {
	// SearXNG design verified case: [{"a":1},{"a":2}] / "a" → [1,2]
	data := mustUnmarshal(t, `[{"a":1},{"a":2}]`)
	results := jsonQuery(data, "a")
	assert.Equal(t, []interface{}{float64(1), float64(2)}, results)
}

func TestJSONQuery_NestedObjects(t *testing.T) {
	data := mustUnmarshal(t, `{"x":{"a":1},"y":{"a":2}}`)
	results := jsonQuery(data, "a")
	assert.Len(t, results, 2)
}

func TestJSONQuery_DeepNesting(t *testing.T) {
	data := mustUnmarshal(t, `{"response":{"results":[{"url":"u1"},{"url":"u2"}]}}`)
	results := jsonQuery(data, "response/results/url")
	assert.Equal(t, []interface{}{"u1", "u2"}, results)
}

func TestJSONQuery_NoMatch(t *testing.T) {
	data := mustUnmarshal(t, `{"a":1}`)
	results := jsonQuery(data, "nonexistent")
	assert.Empty(t, results)
}

func TestJSONQuery_EmptyQuery(t *testing.T) {
	data := mustUnmarshal(t, `{"a":1}`)
	results := jsonQuery(data, "")
	assert.Empty(t, results)
}

func TestJSONQuery_ScalarValue(t *testing.T) {
	data := mustUnmarshal(t, `{"title":"Hello"}`)
	results := jsonQuery(data, "title")
	assert.Equal(t, []interface{}{"Hello"}, results)
}
