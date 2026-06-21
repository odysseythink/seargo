package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestStartpageParse(t *testing.T) {
	resp := `["golang",["golang tutorial","golang playground"]]`
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
	if results[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", results[0])
	}
}
