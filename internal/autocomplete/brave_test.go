package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestBraveParse(t *testing.T) {
	resp := `{"results":[{"query":"golang tutorial"},{"query":"golang playground"}]}`
	var data struct {
		Results []struct {
			Query string `json:"query"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	results := make([]string, 0, len(data.Results))
	for _, r := range data.Results {
		if r.Query != "" {
			results = append(results, r.Query)
		}
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}
	if results[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", results[0])
	}
}

func TestBraveParseEmpty(t *testing.T) {
	resp := `{"results":[]}`
	var data struct {
		Results []struct {
			Query string `json:"query"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Results) != 0 {
		t.Fatalf("expected empty results, got %d", len(data.Results))
	}
}
