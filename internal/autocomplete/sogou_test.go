package autocomplete

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSogouParse(t *testing.T) {
	// Sogou API returns a JSON array: [query, [[suggestion, score], ...], ...]
	// It may have callback padding like: searxng([...], -1)
	resp := `searxng(["golang",[["golang tutorial",1],["golang playground",2]]],-1)`
	body := resp
	start := strings.Index(body, "[")
	end := strings.LastIndex(body, "]")
	rawJSON := body[start : end+1]

	var data []interface{}
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) < 2 {
		t.Fatal("expected at least 2 elements")
	}
	suggestions, ok := data[1].([]interface{})
	if !ok {
		t.Fatal("expected data[1] to be an array")
	}
	results := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		item, ok := s.([]interface{})
		if !ok || len(item) == 0 {
			continue
		}
		if text, ok := item[0].(string); ok && text != "" {
			results = append(results, text)
		}
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}
	if results[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", results[0])
	}
}

func TestSogouParseEmpty(t *testing.T) {
	resp := `searxng(["golang",[],[]],-1)`
	body := resp
	start := strings.Index(body, "[")
	end := strings.LastIndex(body, "]")
	rawJSON := body[start : end+1]

	var data []interface{}
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) < 2 {
		t.Fatal("expected at least 2 elements")
	}
	suggestions, ok := data[1].([]interface{})
	if !ok {
		t.Fatal("expected data[1] to be an array")
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected empty suggestions, got %d", len(suggestions))
	}
}

func TestSogouParseNoCallback(t *testing.T) {
	// Plain JSON array without callback padding
	resp := `["golang",[["golang tutorial",1],["golang playground",2]]]`
	body := resp
	start := strings.Index(body, "[")
	end := strings.LastIndex(body, "]")
	rawJSON := body[start : end+1]

	var data []interface{}
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) < 2 {
		t.Fatal("expected at least 2 elements")
	}
	suggestions, ok := data[1].([]interface{})
	if !ok {
		t.Fatal("expected data[1] to be an array")
	}
	results := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		item, ok := s.([]interface{})
		if !ok || len(item) == 0 {
			continue
		}
		if text, ok := item[0].(string); ok && text != "" {
			results = append(results, text)
		}
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}
}
