package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestWikipediaParse(t *testing.T) {
	resp := `["golang",["Go (programming language)","Golang playground"],["",""],["https://en.wikipedia.org/wiki/Go_(programming_language)",""]]`
	var data []interface{}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) < 2 {
		t.Fatal("expected at least 2 elements")
	}
	suggestions := data[1].([]interface{})
	results := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		if text, ok := s.(string); ok && text != "" {
			results = append(results, text)
		}
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}
	if !(results[0] == "Go (programming language)" || results[0] == "Golang playground") {
		t.Errorf("unexpected result %q", results[0])
	}
}

func TestWikipediaParseEmpty(t *testing.T) {
	resp := `["nonexistent",[]]`
	var data []interface{}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	suggestions := data[1].([]interface{})
	if len(suggestions) != 0 {
		t.Fatalf("expected empty, got %d", len(suggestions))
	}
}
