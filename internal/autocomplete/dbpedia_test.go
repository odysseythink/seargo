package autocomplete

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestDBpediaParse(t *testing.T) {
	resp := `<ArrayOfResults><Result><Label>Go (programming language)</Label></Result><Result><Label>Golang playground</Label></Result></ArrayOfResults>`

	type Result struct {
		Label string `xml:"Label"`
	}
	type ArrayOfResults struct {
		XMLName xml.Name `xml:"ArrayOfResults"`
		Results []Result `xml:"Result"`
	}

	var arr ArrayOfResults
	if err := xml.Unmarshal([]byte(resp), &arr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	results := make([]string, 0, len(arr.Results))
	for _, r := range arr.Results {
		label := strings.TrimSpace(r.Label)
		if label != "" {
			results = append(results, label)
		}
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}
	if results[0] != "Go (programming language)" {
		t.Errorf("expected 'Go (programming language)', got %q", results[0])
	}
}

func TestDBpediaParseEmpty(t *testing.T) {
	resp := `<ArrayOfResults></ArrayOfResults>`

	type Result struct {
		Label string `xml:"Label"`
	}
	type ArrayOfResults struct {
		XMLName xml.Name `xml:"ArrayOfResults"`
		Results []Result `xml:"Result"`
	}

	var arr ArrayOfResults
	if err := xml.Unmarshal([]byte(resp), &arr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(arr.Results) != 0 {
		t.Fatalf("expected empty, got %d", len(arr.Results))
	}
}
