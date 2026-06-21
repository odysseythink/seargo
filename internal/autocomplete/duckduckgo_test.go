package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestDuckDuckGoParse(t *testing.T) {
	resp := `["golang",["golang tutorial","golang playground","golang map"]]`
	var data []interface{}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	suggestions := data[1].([]interface{})

	results := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		if text, ok := s.(string); ok && text != "" {
			results = append(results, text)
		}
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}
	if results[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", results[0])
	}
}

func TestDuckDuckGoParseEmpty(t *testing.T) {
	resp := `["golang",[]]`
	var data []interface{}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	suggestions := data[1].([]interface{})

	if len(suggestions) != 0 {
		t.Fatalf("expected empty suggestions, got %d", len(suggestions))
	}
}
