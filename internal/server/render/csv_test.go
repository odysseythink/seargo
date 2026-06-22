package render

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/seargo/seargo/pkg/models"
)

func TestCSVWriterHeaders(t *testing.T) {
	resp := &models.Response{
		Query:   "search",
		Results: []models.Result{},
	}
	w := &CSVWriter{}
	data, err := w.Render(resp)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	str := string(data)
	for _, col := range []string{"title", "url", "content", "engine", "score", "published_at"} {
		if !strings.Contains(str, col) {
			t.Errorf("missing header column %q in output: %s", col, str)
		}
	}
}

func TestCSVWriterRows(t *testing.T) {
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	resp := &models.Response{
		Query: "golang",
		Results: []models.Result{
			{Title: "The Go Programming Language", URL: "https://go.dev", Content: "Go is an open source programming language", Engine: "google", Score: 0.99, PublishedAt: &ts},
			{Title: "Go Wiki", URL: "https://github.com/golang/go/wiki", Content: "Community wiki for Go", Engine: "bing", Score: 0.87, PublishedAt: nil},
		},
	}
	w := &CSVWriter{}
	data, err := w.Render(resp)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	str := string(data)
	lines := strings.Split(strings.TrimSpace(str), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (header + 2 rows), got %d: %s", len(lines), str)
	}
	reader := csv.NewReader(strings.NewReader(str))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV parse error: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	if records[1][0] != "The Go Programming Language" {
		t.Errorf("title mismatch: %q", records[1][0])
	}
	if records[1][1] != "https://go.dev" {
		t.Errorf("url mismatch: %q", records[1][1])
	}
	if records[1][3] != "google" {
		t.Errorf("engine mismatch: %q", records[1][3])
	}
}

func TestCSVWriterFormulaInjection(t *testing.T) {
	tests := []struct {
		title    string
		expectOK bool
	}{
		{"=CMD|' /C calc'!A0", false},
		{"+1+1", false},
		{"-2+3", false},
		{"@SUM(A1:A10)", false},
		{"normal text", true},
		{"1234", true},
	}
	for _, tc := range tests {
		resp := &models.Response{
			Query: "test",
			Results: []models.Result{
				{Title: tc.title, URL: "https://x.com"},
			},
		}
		w := &CSVWriter{}
		data, err := w.Render(resp)
		if err != nil {
			t.Fatalf("Render error for %q: %v", tc.title, err)
		}
		reader := csv.NewReader(strings.NewReader(string(data)))
		records, _ := reader.ReadAll()
		cell := records[1][0]
		if cell == "" {
			t.Errorf("cell for %q is empty after sanitization — data loss", tc.title)
		}
		if !tc.expectOK {
			if strings.HasPrefix(cell, "=") || strings.HasPrefix(cell, "+") ||
				strings.HasPrefix(cell, "-") || strings.HasPrefix(cell, "@") {
				t.Errorf("cell for %q starts with formula char: %q", tc.title, cell)
			}
		}
	}
}

func TestCSVWriterContentType(t *testing.T) {
	w := &CSVWriter{}
	if w.ContentType() != "text/csv; charset=utf-8" {
		t.Errorf("unexpected content type: %s", w.ContentType())
	}
}
