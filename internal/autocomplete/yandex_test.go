package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestYandexParse(t *testing.T) {
	resp := `["golang",["golang tutorial","golang playground"]]`
	var data []interface{}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
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
		if text, ok := s.(string); ok && text != "" {
			results = append(results, text)
		}
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", results[0])
	}
}
