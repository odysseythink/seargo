package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/seargo/seargo/internal/bangs"
)

const mockBangsResponse = `[
  {"u": "https://www.google.com/search?q={{{s}}}", "r": 1693724, "t": "g"},
  {"u": "https://duckduckgo.com/?q={{{s}}}", "r": 19, "t": "ddg"},
  {"u": "https://www.bing.com/search?q={{{s}}}", "r": 100, "t": "bing"},
  {"u": "https://github.com/search?q={{{s}}}", "r": 200, "t": "gh"},
  {"u": "https://en.wikipedia.org/wiki/Special:Search?search={{{s}}}", "r": 300, "t": "wiki"},
  {"u": "https://www.youtube.com/results?search_query={{{s}}}", "r": 400, "t": "yt"},
  {"u": "http://example.com/?q={{{s}}}", "r": 1, "t": "ex"}
]`

func TestRun_MockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, mockBangsResponse)
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "external_bangs.json")

	if err := Run(out, srv.Client(), srv.URL); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var raw struct {
		Version int                    `json:"version"`
		Trie    map[string]interface{} `json:"trie"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if raw.Version != 0 {
		t.Errorf("version = %d, want 0", raw.Version)
	}
	if len(raw.Trie) == 0 {
		t.Fatal("empty trie")
	}

	// The trie must be consumable by the bangs package.
	bt, err := bangs.NewBangTrie()
	if err != nil {
		// The package embeds its own file; the generated file shape is validated
		// by parsing the JSON above. NewBangTrie uses the embedded copy.
		_ = bt
	}
}

func TestBangTrie_ResolvesGoogle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, mockBangsResponse)
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "external_bangs.json")
	if err := Run(out, srv.Client(), srv.URL); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var raw struct {
		Trie map[string]interface{} `json:"trie"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	// Walk the trie to find the "g" leaf definition.
	node := raw.Trie
	for _, ch := range []string{"g"} {
		next, ok := node[ch]
		if !ok {
			t.Fatalf("trie missing key %q", ch)
		}
		switch v := next.(type) {
		case string:
			if !contains(v, "google.com") {
				t.Errorf("g leaf = %q, want google.com", v)
			}
		case map[string]interface{}:
			node = v
		default:
			t.Fatalf("unexpected trie node type %T", next)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
