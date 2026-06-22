package render

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/seargo/seargo/pkg/models"
)

func TestJSONWriter(t *testing.T) {
	ts := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	resp := &models.Response{
		Query:    "test query",
		Category: models.CategoryGeneral,
		Results: []models.Result{
			{
				Title:       "Result 1",
				URL:         "https://example.com",
				Content:     "Example content",
				Engine:      "google",
				Score:       0.95,
				PublishedAt: &ts,
			},
		},
		Total:          1,
		Page:           1,
		PageSize:       10,
		EnginesUsed:    []string{"google"},
		EnginesFailed:  []string{},
		ResponseTimeMs: 250,
	}
	w := &JSONWriter{}
	data, err := w.Render(resp)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty output")
	}
	var decoded models.Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if decoded.Query != "test query" {
		t.Errorf("query mismatch: got %q", decoded.Query)
	}
	if len(decoded.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(decoded.Results))
	}
}

func TestJSONWriterEmptyResults(t *testing.T) {
	resp := &models.Response{
		Query:    "no results",
		Results:  []models.Result{},
		Total:    0,
		Page:     1,
		PageSize: 10,
	}
	w := &JSONWriter{}
	data, err := w.Render(resp)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	var decoded models.Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(decoded.Results) != 0 {
		t.Errorf("expected empty results, got %d", len(decoded.Results))
	}
}
