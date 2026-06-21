package autocomplete

import (
	"encoding/json"
	"testing"
)

func TestQwantParse(t *testing.T) {
	resp := `{"status":"success","data":{"items":[{"value":"golang tutorial"},{"value":"golang playground"}]}}`
	var data struct {
		Status string `json:"status"`
		Data   struct {
			Items []struct {
				Value string `json:"value"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Status != "success" {
		t.Fatal("expected success status")
	}
	if len(data.Data.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(data.Data.Items))
	}
	if data.Data.Items[0].Value != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", data.Data.Items[0].Value)
	}
}

func TestQwantParse_NonSuccess(t *testing.T) {
	resp := `{"status":"error","data":{"items":[]}}`
	var data struct {
		Status string `json:"status"`
		Data   struct {
			Items []struct {
				Value string `json:"value"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Status != "error" {
		t.Fatal("expected error status")
	}
}
