package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestSwisscowsParse(t *testing.T) {
	resp := `["golang tutorial","golang playground","golang map"]`
	var data []string
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(data), data)
	}
	if data[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", data[0])
	}
}

func TestSwisscowsParseEmpty(t *testing.T) {
	resp := `[]`
	var data []string
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty, got %d", len(data))
	}
}
