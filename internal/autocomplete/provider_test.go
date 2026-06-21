package autocomplete

import (
	"context"
	"testing"
	"time"

	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/config"
)

type stubProvider struct {
	results []string
	err     error
}

func (s *stubProvider) Fetch(_ context.Context, _ string, _ string) ([]string, error) {
	return s.results, s.err
}


func TestRegisterDuplicatePanics(t *testing.T) {
	Reset()
	Register("test", &stubProvider{})
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	Register("test", &stubProvider{})
}

func TestRegisterGetAllNames(t *testing.T) {
	Reset()
	Register("a", &stubProvider{})
	Register("b", &stubProvider{})

	p, ok := Get("a")
	if !ok || p == nil {
		t.Fatal("expected to get provider 'a'")
	}
	_, ok = Get("nonexistent")
	if ok {
		t.Fatal("expected false for nonexistent provider")
	}

	all := All()
	if len(all) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(all))
	}

	names := Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}

	Reset()
	if len(All()) != 0 {
		t.Fatal("expected empty after reset")
	}
}

func TestCacheSetGetExpiry(t *testing.T) {
	c := NewResultCache(makeTestKV(t).WithNamespace("autocomplete"), 100 * time.Millisecond)
	defer c.Close()

	c.Set("k1", []string{"a", "b"})
	got, ok := c.Get("k1")
	if !ok || len(got) != 2 || got[0] != "a" {
		t.Fatalf("expected [a b], got %v (ok=%v)", got, ok)
	}

	_, ok = c.Get("nonexistent")
	if ok {
		t.Fatal("expected miss for nonexistent key")
	}

	time.Sleep(150 * time.Millisecond)
	_, ok = c.Get("k1")
	if ok {
		t.Fatal("expected miss after TTL expiry")
	}
}

func TestServiceUnknownBackend(t *testing.T) {
	Reset()
	client := newHttpxTestClient()
	svc := NewService(client, NewResultCache(makeTestKV(t).WithNamespace("autocomplete"), 10*time.Second))
	defer svc.Cache().Close()

	result := svc.Suggest(context.Background(), "nonexistent", "test", "en-US")
	if len(result) != 0 {
		t.Fatalf("expected empty for unknown backend, got %v", result)
	}
}

func TestServiceShortQuery(t *testing.T) {
	Reset()
	Register("test", &stubProvider{results: []string{"should not appear"}})
	client := newHttpxTestClient()
	svc := NewService(client, NewResultCache(makeTestKV(t).WithNamespace("autocomplete"), 10*time.Second))
	defer svc.Cache().Close()

	result := svc.Suggest(context.Background(), "test", "x", "en-US")
	if len(result) != 0 {
		t.Fatalf("expected empty for short query, got %v", result)
	}
}

func TestServicePanicRecovery(t *testing.T) {
	Reset()
	Register("panic", &stubProvider{})
	client := newHttpxTestClient()
	svc := NewService(client, NewResultCache(makeTestKV(t).WithNamespace("autocomplete"), 10*time.Second))
	defer svc.Cache().Close()

	// Override the provider to make it panic
	providersMu.Lock()
	providers["panic"] = &panicProvider{}
	providersMu.Unlock()

	result := svc.Suggest(context.Background(), "panic", "test", "en-US")
	if len(result) != 0 {
		t.Fatalf("expected empty after panic recovery, got %v", result)
	}
}

type panicProvider struct{}

func (p *panicProvider) Fetch(_ context.Context, _ string, _ string) ([]string, error) {
	panic("intentional panic")
}

func TestLocaleHelpers(t *testing.T) {
	if LocaleToLanguage("en-US") != "en" {
		t.Fatalf("expected 'en', got %q", LocaleToLanguage("en-US"))
	}
	if LocaleToLanguage("zh_CN") != "zh" {
		t.Fatalf("expected 'zh', got %q", LocaleToLanguage("zh_CN"))
	}
	if LocaleToLanguage("fr") != "fr" {
		t.Fatalf("expected 'fr', got %q", LocaleToLanguage("fr"))
	}
	if LocaleToCountry("en-US") != "US" {
		t.Fatalf("expected 'US', got %q", LocaleToCountry("en-US"))
	}
	if LocaleToCountry("fr") != "" {
		t.Fatalf("expected empty, got %q", LocaleToCountry("fr"))
	}
	if LocaleToDDGRegion("en-US") != "us-en" {
		t.Fatalf("expected 'us-en', got %q", LocaleToDDGRegion("en-US"))
	}
	if LocaleToDDGRegion("fr") != "fr-fr" {
		t.Fatalf("expected 'fr-fr', got %q", LocaleToDDGRegion("fr"))
	}
	if LocaleToGoogleSubdomain("de") != "www.google.de" {
		t.Fatalf("expected www.google.de, got %q", LocaleToGoogleSubdomain("de"))
	}
	if LocaleToGoogleSubdomain("xx") != "www.google.com" {
		t.Fatalf("expected www.google.com, got %q", LocaleToGoogleSubdomain("xx"))
	}
	if LocaleToStartpageLanguage("de") != "deutsch" {
		t.Fatalf("expected deutsch, got %q", LocaleToStartpageLanguage("de"))
	}
	if LocaleToWikipediaLang("de") != "de" {
		t.Fatalf("expected de, got %q", LocaleToWikipediaLang("de"))
	}
	if LocaleToWikipediaNetloc("de") != "de.wikipedia.org" {
		t.Fatalf("expected de.wikipedia.org, got %q", LocaleToWikipediaNetloc("de"))
	}
}

func TestCacheClose(t *testing.T) {
	c := NewResultCache(makeTestKV(t).WithNamespace("autocomplete"), time.Second)
	c.Close()
	// Closing twice should not panic
	c.Close()
}

// newHttpxTestClient creates an httpx.Client for tests.
func newHttpxTestClient() *httpx.Client {
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout: 10,
			MaxRedirects:   30,
			EnableHTTP:     true,
		},
	}
	reg, err := httpx.NewRegistry(cfg)
	if err != nil {
		panic("failed to create test registry: " + err.Error())
	}
	return httpx.NewClient(reg, "", "", "SearGoTest/1.0", 10*time.Second)
}
