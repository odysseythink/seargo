package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestQuarkParse(t *testing.T) {
	resp := `{"data":[{"topic":"golang tutorial"},{"topic":"golang playground"}]}`
	var data struct {
		Data []struct {
			Topic string `json:"topic"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	results := make([]string, 0, len(data.Data))
	for _, d := range data.Data {
		if d.Topic != "" {
			results = append(results, d.Topic)
		}
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", results[0])
	}
}
