package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestSeznamParse(t *testing.T) {
	resp := `["golang",[["golang tutorial",100],["golang playground",50]],{"a":""}]`
	var data []interface{}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) < 2 {
		t.Fatal("expected at least 2 elements")
	}
	phrases := data[1].([]interface{})
	results := make([]string, 0, len(phrases))
	for _, p := range phrases {
		item := p.([]interface{})
		if text, ok := item[0].(string); ok && text != "" {
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
