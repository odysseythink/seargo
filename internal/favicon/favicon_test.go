package favicon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

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
