package bangs

import (
	"strings"
	"testing"
)

func TestParseDef(t *testing.T) {
	tests := []struct {
		raw      string
		wantURL  string
		wantRank int
	}{
		{"//www.google.com/search?q=\x02\x011693724",
			"//www.google.com/search?q=\x02", 1693724},
		{"http://duckduckgo.com/?q=\x02\x0119",
			"http://duckduckgo.com/?q=\x02", 19},
		{"https://example.com/?q=\x02\x010",
			"https://example.com/?q=\x02", 0},
	}
	for _, tc := range tests {
		def := parseDef(tc.raw)
		if def.URL != tc.wantURL {
			t.Errorf("parseDef URL = %q, want %q", def.URL, tc.wantURL)
		}
		if def.Rank != tc.wantRank {
			t.Errorf("parseDef Rank = %d, want %d", def.Rank, tc.wantRank)
		}
	}
}

func TestNewBangTrie_LoadsSuccessfully(t *testing.T) {
	bt, err := NewBangTrie()
	if err != nil {
		t.Fatalf("NewBangTrie failed: %v", err)
	}
	if bt == nil || bt.root == nil || len(bt.root) == 0 {
		t.Fatal("empty trie root")
	}
}

func TestResolve_KnownBangs(t *testing.T) {
	bt, err := NewBangTrie()
	if err != nil {
		t.Fatalf("NewBangTrie: %v", err)
	}
	tests := []struct {
		bang, query, wantContain string
	}{
		{"g", "test", "google.com/search"},
		{"ddg", "hello", "duckduckgo.com"},
		{"bing", "world", "bing.com/search"},
		{"gh", "repo", "github.com/search"},
		{"wiki", "go", "wikipedia.org"},
		{"yt", "video", "youtube.com"},
	}
	for _, tc := range tests {
		got := bt.Resolve(tc.bang, tc.query)
		if got == nil {
			t.Errorf("Resolve(%q,%q) = nil", tc.bang, tc.query)
			continue
		}
		if !strings.Contains(*got, tc.wantContain) {
			t.Errorf("Resolve(%q,%q) = %q, want containing %q",
				tc.bang, tc.query, *got, tc.wantContain)
		}
	}
}

func TestResolve_EmptyQueryReturnsRootDomain(t *testing.T) {
	bt, _ := NewBangTrie()
	got := bt.Resolve("g", "")
	if got == nil {
		t.Fatal("Resolve('g','') = nil")
	}
	if !strings.HasSuffix(*got, "/") || strings.Contains(*got, "search") {
		t.Errorf("Resolve('g','') = %q, want root domain with trailing slash", *got)
	}
}

func TestResolve_UnknownBang(t *testing.T) {
	bt, _ := NewBangTrie()
	// A prefix with no match at all should return nil.
	if got := bt.Resolve("NONEXISTENT", "test"); got != nil {
		t.Errorf("expected nil for completely unknown bang, got %q", *got)
	}
	// Partial matches return the last found definition along the prefix path.
	// "zzzzz..." matches root 'z' and its child 'z', so a partial result is expected.
	got := bt.Resolve("zzzzz_nonexistent_bang", "test")
	if got == nil {
		t.Fatal("expected partial match result for prefix with known root, got nil")
	}
}

func TestSuggest(t *testing.T) {
	bt, _ := NewBangTrie()
	suggestions := bt.Suggest("gi")
	if len(suggestions) == 0 {
		t.Fatal("expected non-empty suggestions for 'gi'")
	}
	found := false
	for _, s := range suggestions {
		if s == "github" || s == "gist" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'github'/'gist' in suggestions for 'gi', got %v", suggestions)
	}
}

func TestSuggest_UnknownPrefix(t *testing.T) {
	bt, _ := NewBangTrie()
	if suggestions := bt.Suggest("zzzzz_nonexistent"); len(suggestions) != 0 {
		t.Errorf("expected empty for unknown prefix, got %v", suggestions)
	}
}

func TestRootDomain(t *testing.T) {
	bt := &BangTrie{root: make(map[string]interface{})}
	tests := []struct{ template, expected string }{
		{"//www.google.com/search?q=\x02", "https://www.google.com/"},
		{"http://example.com/path", "http://example.com/"},
		{"https://test.org/x?p=1", "https://test.org/"},
	}
	for _, tc := range tests {
		got := bt.rootDomain(tc.template)
		if got == nil {
			t.Errorf("rootDomain(%q) = nil", tc.template)
			continue
		}
		if *got != tc.expected {
			t.Errorf("rootDomain(%q) = %q, want %q", tc.template, *got, tc.expected)
		}
	}
}
