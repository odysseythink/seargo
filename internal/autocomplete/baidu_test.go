package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestBaiduParse(t *testing.T) {
	resp := `{"s":["golang tutorial","golang playground"]}`
	var data struct {
		S []string `json:"s"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.S) != 2 {
		t.Fatalf("expected 2 results, got %d", len(data.S))
	}
	if data.S[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", data.S[0])
	}
}
