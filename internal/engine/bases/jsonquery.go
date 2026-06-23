package bases

import (
	"strings"
)

// jsonQuery traverses arbitrary JSON data using a slash-delimited path,
// collecting leaf values at each step. Arrays and objects are descended
// automatically. This implements the SearXNG json_engine.query semantics.
//
// Algorithm (from design):
// 1. Split query by "/".
// 2. Recursively traverse data: match current key, descend into iterables.
// 3. Return all matched leaf values.
func jsonQuery(data interface{}, query string) []interface{} {
	if query == "" {
		return nil
	}

	parts := strings.Split(query, "/")
	return queryRecursive(data, parts)
}

func queryRecursive(data interface{}, parts []string) []interface{} {
	if len(parts) == 0 {
		return nil
	}

	current := parts[0]
	remaining := parts[1:]

	// 支持 XML-to-JSON 常见的 "$" 文本节点键。
	if current == "$" {
		switch v := data.(type) {
		case map[string]interface{}:
			if val, ok := v["$"]; ok {
				return collectValue(val, remaining)
			}
		}
		return nil
	}

	var results []interface{}

	switch v := data.(type) {
	case map[string]interface{}:
		if val, ok := v[current]; ok {
			results = append(results, collectValue(val, remaining)...)
		}
		// Also search nested objects (for query "a" matching {"x":{"a":1}})
		for _, val := range v {
			results = append(results, queryRecursive(val, parts)...)
		}

	case []interface{}:
		for _, item := range v {
			results = append(results, queryRecursive(item, parts)...)
		}
	}

	return results
}

// collectValue collects leaf values from a matched intermediate node.
// If there are more path parts, continues traversal; otherwise returns
// the value itself (wrapped in slice).
func collectValue(data interface{}, remaining []string) []interface{} {
	if len(remaining) == 0 {
		// Flatten slices so a path ending at an array (e.g. "response/docs")
		// yields individual elements, not the array wrapped in a single element.
		if arr, ok := data.([]interface{}); ok {
			return arr
		}
		return []interface{}{data}
	}

	next := remaining[0]
	rest := remaining[1:]
	if next == "$" {
		switch v := data.(type) {
		case map[string]interface{}:
			if val, ok := v["$"]; ok {
				return collectValue(val, rest)
			}
		}
		return nil
	}

	var results []interface{}

	switch v := data.(type) {
	case map[string]interface{}:
		return queryRecursive(v, remaining)
	case []interface{}:
		for _, item := range v {
			results = append(results, collectValue(item, remaining)...)
		}
		return results
	default:
		// Scalar with remaining path parts → no match
		return nil
	}
}
