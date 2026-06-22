package render

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/seargo/seargo/pkg/models"
)

func TestRSSWriterStructure(t *testing.T) {
	ts := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	resp := &models.Response{
		Query: "golang tutorial",
		Results: []models.Result{
			{Title: "Go Tutorial", URL: "https://example.com/go", Content: "Learn Go programming", Engine: "google", PublishedAt: &ts},
		},
	}
	w := &RSSWriter{}
	data, err := w.Render(resp, "https://seargo.example.com")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	str := string(data)
	if !strings.Contains(str, "<feed") {
		t.Fatal("output missing <feed> element")
	}
	if !strings.Contains(str, "<entry") {
		t.Fatal("output missing <entry> element")
	}
	if !strings.Contains(str, "golang tutorial") {
		t.Error("output missing query in title")
	}
	if !strings.Contains(str, "Go Tutorial") {
		t.Error("output missing result title")
	}
	var feed struct {
		XMLName xml.Name `xml:"feed"`
		Title   string   `xml:"title"`
		Entry   []struct {
			Title string `xml:"title"`
			Link  struct{ Href string `xml:"href,attr"` } `xml:"link"`
		} `xml:"entry"`
	}
	if err := xml.Unmarshal(data, &feed); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}
	if len(feed.Entry) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(feed.Entry))
	}
}

func TestRSSWriterMultipleEntries(t *testing.T) {
	resp := &models.Response{
		Query: "test",
		Results: []models.Result{
			{Title: "A", URL: "https://a.com", Engine: "e1"},
			{Title: "B", URL: "https://b.com", Engine: "e2"},
			{Title: "C", URL: "https://c.com", Engine: "e3"},
		},
	}
	w := &RSSWriter{}
	data, err := w.Render(resp, "https://seargo.example.com")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	var feed struct{ Entry []struct{} `xml:"entry"` }
	if err := xml.Unmarshal(data, &feed); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}
	if len(feed.Entry) != 3 {
		t.Errorf("expected 3 entries, got %d", len(feed.Entry))
	}
}

func TestRSSWriterXMLEscape(t *testing.T) {
	resp := &models.Response{
		Query: "test",
		Results: []models.Result{
			{Title: "A & B < C > D", URL: "https://x.com", Content: "text with \"quotes\" & ampersand", Engine: "google"},
		},
	}
	w := &RSSWriter{}
	data, err := w.Render(resp, "https://seargo.example.com")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	var feed struct {
		Entry []struct {
			Title   string `xml:"title"`
			Summary string `xml:"summary"`
		} `xml:"entry"`
	}
	if err := xml.Unmarshal(data, &feed); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}
	if feed.Entry[0].Title != "A & B < C > D" {
		t.Errorf("title not round-tripped: %q", feed.Entry[0].Title)
	}
	if feed.Entry[0].Summary != `text with "quotes" & ampersand` {
		t.Errorf("summary not round-tripped: %q", feed.Entry[0].Summary)
	}
}

func TestRSSWriterEmptyResults(t *testing.T) {
	resp := &models.Response{Query: "empty", Results: []models.Result{}}
	w := &RSSWriter{}
	data, err := w.Render(resp, "https://seargo.example.com")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	var feed struct{ Entry []struct{} `xml:"entry"` }
	if err := xml.Unmarshal(data, &feed); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}
	if len(feed.Entry) != 0 {
		t.Errorf("expected 0 entries, got %d", len(feed.Entry))
	}
}

func TestRSSWriterContentType(t *testing.T) {
	w := &RSSWriter{}
	if w.ContentType() != "application/atom+xml; charset=utf-8" {
		t.Errorf("unexpected content type: %s", w.ContentType())
	}
}
