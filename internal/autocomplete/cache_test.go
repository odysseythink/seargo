package autocomplete

import (
	"testing"
	"time"

	"github.com/seargo/seargo/internal/storage"
)

func makeTestKV(t *testing.T) storage.KV {
	t.Helper()
	kv, err := storage.New(storage.Options{
		Backend:     "memory",
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { kv.Close() })
	return kv
}

func TestResultCache_RoundTrip(t *testing.T) {
	kv := makeTestKV(t)
	c := NewResultCache(kv.WithNamespace("autocomplete"), DefaultCacheTTL)
	defer c.Close()

	c.Set("key1", []string{"a", "b", "c"})
	results, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected hit")
	}
	if len(results) != 3 || results[0] != "a" || results[1] != "b" || results[2] != "c" {
		t.Fatalf("got %v, want [a b c]", results)
	}
}

func TestResultCache_TTLExpiry(t *testing.T) {
	kv := makeTestKV(t)
	c := NewResultCache(kv.WithNamespace("ac"), 10*time.Millisecond)
	defer c.Close()

	c.Set("temp", []string{"x"})
	time.Sleep(50 * time.Millisecond)

	_, ok := c.Get("temp")
	if ok {
		t.Fatal("expected expiry")
	}
}

func TestResultCache_Miss(t *testing.T) {
	kv := makeTestKV(t)
	c := NewResultCache(kv.WithNamespace("ac"), DefaultCacheTTL)
	defer c.Close()

	_, ok := c.Get("no_such_key")
	if ok {
		t.Fatal("expected miss")
	}
}

func TestResultCache_DefaultTTL(t *testing.T) {
	kv := makeTestKV(t)
	c := NewResultCache(kv.WithNamespace("ac"), 0)
	defer c.Close()

	c.Set("k", []string{"v"})
	results, ok := c.Get("k")
	if !ok || len(results) != 1 || results[0] != "v" {
		t.Fatalf("default TTL: ok=%v results=%v", ok, results)
	}
}
