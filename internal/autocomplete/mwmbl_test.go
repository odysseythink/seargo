package autocomplete

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMwmblParse(t *testing.T) {
	resp := `{"suggestions":["golang tutorial","golang playground","python tutorial"]}`
	var data struct {
		Suggestions []string `json:"suggestions"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	queryLower := strings.ToLower("golang")
	results := make([]string, 0, len(data.Suggestions))
	for _, s := range data.Suggestions {
		if strings.HasPrefix(strings.ToLower(s), queryLower) && s != "" {
			results = append(results, s)
		}
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results after prefix filter, got %d: %v", len(results), results)
	}
}

func TestMwmblParseEmpty(t *testing.T) {
	resp := `{"suggestions":[]}`
	var data struct {
		Suggestions []string `json:"suggestions"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Suggestions) != 0 {
		t.Fatalf("expected empty, got %d", len(data.Suggestions))
	}
}
