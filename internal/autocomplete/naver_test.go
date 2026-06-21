package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestNaverParse(t *testing.T) {
	resp := `[["golang"],["golang tutorial","golang playground"]]`
	var data [][]string
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) < 2 {
		t.Fatal("expected at least 2 elements")
	}
	results := data[1]
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", results[0])
	}
}
