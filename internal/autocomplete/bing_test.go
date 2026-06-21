package autocomplete

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBingParse_PUAStrip(t *testing.T) {
	resp := `{"s":[{"q":"golang \ue000tutorial\ue001"},{"q":"\ue000go\ue001lang playground"},{"q":"plain result"}]}`

	var data struct {
		S []struct {
			Q string `json:"q"`
		} `json:"s"`
	}
	if err := json.Unmarshal([]byte(resp), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var results []string
	for _, s := range data.S {
		text := s.Q
		text = strings.ReplaceAll(text, bingPUASpan, "")
		text = strings.ReplaceAll(text, bingPUAEnd, "")
		text = strings.TrimSpace(text)
		if text != "" {
			results = append(results, text)
		}
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}
	if results[0] != "golang tutorial" {
		t.Errorf("expected 'golang tutorial', got %q", results[0])
	}
	if results[1] != "golang playground" {
		t.Errorf("expected 'golang playground', got %q", results[1])
	}
	if results[2] != "plain result" {
		t.Errorf("expected 'plain result' to survive, got %q", results[2])
	}
}

func TestRandomCVID(t *testing.T) {
	c1 := randomCVID()
	c2 := randomCVID()
	if len(c1) != 32 {
		t.Fatalf("expected 32 chars, got %d", len(c1))
	}
	if c1 == c2 {
		t.Fatal("expected different random CVIDs")
	}
	for _, c := range c1 {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Fatalf("unexpected char %c in CVID", c)
		}
	}
}
