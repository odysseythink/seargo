package favicon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/seargo/seargo/internal/security"
	"github.com/seargo/seargo/internal/storage"
)

func makeTestKV(t *testing.T) storage.KV {
	t.Helper()
	kv, err := storage.New(storage.Options{
		Backend:     "memory",
		NumCounters: 10_000,
		MaxCost:     10 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { kv.Close() })
	return kv.WithNamespace("favicon")
}

var testCfg = Config{
	Cache: CacheConfig{
		HoldTime:     time.Hour,
		BlobMaxBytes: 20480,
	},
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func TestSearchFavicon_CacheHit(t *testing.T) {
	kv := makeTestKV(t)
	svc := New(testCfg, nil, kv)

	mime := "image/png"
	blob := []byte{0x89, 0x50, 0x4E, 0x47}
	raw := append([]byte(mime+"\n"), blob...)
	hash := sha256Hex(raw)
	kv.Set(context.Background(), "blob:"+hash, raw, time.Hour)
	kv.Set(context.Background(), "map:testresolver:example.com", []byte(hash), time.Hour)

	data, gotMime, err := svc.SearchFavicon(context.Background(), "testresolver", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(blob) {
		t.Fatalf("blob: got %x, want %x", data, blob)
	}
	if gotMime != mime {
		t.Fatalf("mime: got %q, want %q", gotMime, mime)
	}
}

func TestSearchFavicon_CacheMiss_CallsResolver(t *testing.T) {
	kv := makeTestKV(t)

	calls := 0
	Register("testresolver2", func(ctx context.Context, authority string) ([]byte, string, error) {
		calls++
		return []byte("FAVICONDATA"), "image/png", nil
	})

	svc := New(testCfg, nil, kv)

	data, mime, err := svc.SearchFavicon(context.Background(), "testresolver2", "example.org")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 resolver call, got %d", calls)
	}
	if string(data) != "FAVICONDATA" {
		t.Fatalf("data: got %q", data)
	}
	if mime != "image/png" {
		t.Fatalf("mime: got %q", mime)
	}

	data2, mime2, err := svc.SearchFavicon(context.Background(), "testresolver2", "example.org")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected still 1 resolver call after cache, got %d", calls)
	}
	if string(data2) != "FAVICONDATA" {
		t.Fatalf("cached data: got %q", data2)
	}
	if mime2 != "image/png" {
		t.Fatalf("cached mime: got %q", mime2)
	}
}

func TestSearchFavicon_NegativeCache(t *testing.T) {
	kv := makeTestKV(t)

	calls := 0
	Register("testresolver3", func(ctx context.Context, authority string) ([]byte, string, error) {
		calls++
		return nil, "", fmt.Errorf("not found")
	})

	svc := New(testCfg, nil, kv)

	_, _, err := svc.SearchFavicon(context.Background(), "testresolver3", "example.net")
	if err == nil {
		t.Fatal("expected error from resolver")
	}
	if calls != 1 {
		t.Fatalf("expected 1 resolver call, got %d", calls)
	}

	_, _, err = svc.SearchFavicon(context.Background(), "testresolver3", "example.net")
	if err == nil {
		t.Fatal("expected error from negative cache")
	}
	if calls != 1 {
		t.Fatalf("expected still 1 resolver call after negative cache, got %d", calls)
	}
}

func TestSearchFavicon_BlobTooBig(t *testing.T) {
	kv := makeTestKV(t)

	cfg := Config{
		Cache: CacheConfig{
			HoldTime:     time.Hour,
			BlobMaxBytes: 10, // very small
		},
	}

	calls := 0
	Register("testresolver4", func(ctx context.Context, authority string) ([]byte, string, error) {
		calls++
		data := make([]byte, 100)
		for i := range data {
			data[i] = byte(i)
		}
		return data, "image/png", nil
	})

	svc := New(cfg, nil, kv)

	// First call: resolver succeeds but blob too big → not cached
	data, mime, err := svc.SearchFavicon(context.Background(), "testresolver4", "big.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 resolver call, got %d", calls)
	}
	if mime != "image/png" {
		t.Fatalf("mime: got %q", mime)
	}
	if len(data) != 100 {
		t.Fatalf("data len: got %d, want 100", len(data))
	}

	// Second call: should call resolver again (not cached)
	_, _, err = svc.SearchFavicon(context.Background(), "testresolver4", "big.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 resolver calls (blob not cached), got %d", calls)
	}
}

func TestSignedURL_WithSigner(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	kv := makeTestKV(t)
	svc := New(testCfg, signer, kv)

	signed, err := svc.SignedURL("testresolver", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(signed, "/favicon_proxy?") {
		t.Fatalf("expected /favicon_proxy? prefix, got %q", signed)
	}
	if !strings.Contains(signed, "h=") {
		t.Fatal("signed URL must have h= parameter")
	}
}

func TestSignedURL_WithoutSigner(t *testing.T) {
	kv := makeTestKV(t)
	svc := New(testCfg, nil, kv)

	signed, err := svc.SignedURL("testresolver", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(signed, "/favicon_proxy?") {
		t.Fatalf("expected /favicon_proxy? prefix, got %q", signed)
	}
	if strings.Contains(signed, "h=") {
		t.Fatal("unsigned URL should not have h= parameter")
	}
}

func TestServe_ValidSignature(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	kv := makeTestKV(t)
	svc := New(testCfg, signer, kv)

	Register("testServe", func(ctx context.Context, authority string) ([]byte, string, error) {
		return []byte("FAVICON"), "image/png", nil
	})

	raw := fmt.Sprintf("%s::%s", "testServe", "example.org")
	sig := signer.Sign([]byte(raw))

	data, mime, err := svc.Serve(context.Background(), "testServe", "example.org", sig)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "FAVICON" {
		t.Fatalf("data: got %q, want %q", data, "FAVICON")
	}
	if mime != "image/png" {
		t.Fatalf("mime: got %q, want %q", mime, "image/png")
	}
}

func TestServe_InvalidSignature(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	kv := makeTestKV(t)
	svc := New(testCfg, signer, kv)

	_, _, err := svc.Serve(context.Background(), "testresolver", "example.org", "badsig")
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestServe_MissingParams(t *testing.T) {
	kv := makeTestKV(t)
	svc := New(testCfg, nil, kv)

	_, _, err := svc.Serve(context.Background(), "", "example.org", "")
	if err == nil {
		t.Fatal("expected error for missing resolver")
	}

	_, _, err = svc.Serve(context.Background(), "testresolver", "", "")
	if err == nil {
		t.Fatal("expected error for missing authority")
	}
}

func TestRewriteFaviconURL(t *testing.T) {
	svc := New(testCfg, nil, makeTestKV(t))

	// Empty favicon → unchanged
	if svc.RewriteFaviconURL("https://ex.com/page", "") != "" {
		t.Fatal("empty favicon should return empty")
	}

	// Non-empty favicon returned as-is (rewriting is done by SignedURL in the handler)
	result := svc.RewriteFaviconURL("https://ex.com/page", "https://ex.com/favicon.ico")
	if result != "https://ex.com/favicon.ico" {
		t.Fatalf("rewritten: got %q", result)
	}
}

func TestInitResolvers(t *testing.T) {
	InitResolvers()

	for _, name := range []string{"allesedv", "duckduckgo", "google", "yandex"} {
		fn, err := GetResolver(name)
		if err != nil {
			t.Fatalf("expected resolver %q to be registered after InitResolvers", name)
		}
		_, _, err = fn(context.Background(), "example.com")
		if err == nil {
			t.Fatalf("expected todo resolver %q to return error", name)
		}
	}
}
