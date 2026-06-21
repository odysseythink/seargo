package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestPrivacyWallParse(t *testing.T) {
	resp := `["golang tutorial","golang playground"]`
	var data []string
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 results, got %d", len(data))
	}
	if data[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", data[0])
	}
}
