package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOptionsDefaults(t *testing.T) {
	// Compile-time: Options struct fields exist with correct types.
	opts := Options{
		Backend:     "memory",
		MaxValueLen: 10240,
		NumCounters: 10_000_000,
		MaxCost:     256 << 20,
		BufferItems: 64,
		Maintenance: time.Hour,
	}
	_ = opts
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"search", "search"},
		{"engine:google", "engine:google"},
		{"weird key!", "weird_key_"},
		{"a/b\\c", "a_b_c"},
		{"hello-world_123:test", "hello-world_123:test"},
		{"", ""},
		{"你好世界", "____"},
		{"sp ace", "sp_ace"},
	}
	for _, tt := range tests {
		got := sanitize(tt.input)
		if got != tt.expected {
			t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// --- Memory backend tests ---

func TestMemoryBackend_RoundTrip(t *testing.T) {
	kv, err := newMemoryBackend(Options{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	key := "testkey"
	value := []byte("hello world")

	if err := kv.Set(ctx, key, value, time.Hour); err != nil {
		t.Fatal(err)
	}

	got, ok, err := kv.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected key to exist")
	}
	if string(got) != string(value) {
		t.Fatalf("got %q, want %q", got, value)
	}
}

func TestMemoryBackend_TTLExpiry(t *testing.T) {
	kv, err := newMemoryBackend(Options{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	if err := kv.Set(ctx, "k", []byte("v"), 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	_, ok, err := kv.Get(ctx, "k")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected key to be expired, but got a hit")
	}
}

func TestMemoryBackend_Delete(t *testing.T) {
	kv, err := newMemoryBackend(Options{NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	kv.Set(ctx, "k", []byte("v"), time.Hour)
	kv.Delete(ctx, "k")

	_, ok, _ := kv.Get(ctx, "k")
	if ok {
		t.Fatal("key should be deleted")
	}
}

func TestMemoryBackend_SetNX(t *testing.T) {
	kv, err := newMemoryBackend(Options{NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	set, err := kv.SetNX(ctx, "nxkey", []byte("first"), time.Hour)
	if err != nil || !set {
		t.Fatal("first SetNX should succeed")
	}

	set, err = kv.SetNX(ctx, "nxkey", []byte("second"), time.Hour)
	if err != nil || set {
		t.Fatal("second SetNX should fail")
	}

	got, ok, _ := kv.Get(ctx, "nxkey")
	if !ok || string(got) != "first" {
		t.Fatalf("got %q ok=%v, want 'first'", got, ok)
	}
}

func TestMemoryBackend_Incr(t *testing.T) {
	kv, err := newMemoryBackend(Options{NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	v, err := kv.Incr(ctx, "counter", time.Hour)
	if err != nil || v != 1 {
		t.Fatalf("first Incr: got %d, want 1", v)
	}
	v, err = kv.Incr(ctx, "counter", time.Hour)
	if err != nil || v != 2 {
		t.Fatalf("second Incr: got %d, want 2", v)
	}
}

func TestMemoryBackend_IncrConcurrency(t *testing.T) {
	kv, err := newMemoryBackend(Options{
		NumCounters: 100_000,
		MaxCost:     10 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	n := 50
	incPerGoroutine := 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incPerGoroutine; j++ {
				kv.Incr(ctx, "shared_counter", time.Hour)
			}
		}()
	}
	wg.Wait()

	v, err := kv.Incr(ctx, "shared_counter", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	expected := int64(n*incPerGoroutine + 1) // +1 because we just did one more Incr
	if v != expected {
		t.Fatalf("concurrent Incr: got %d, want %d", v, expected)
	}
}

func TestMemoryBackend_BackendName(t *testing.T) {
	kv, _ := newMemoryBackend(Options{NumCounters: 100, MaxCost: 1 << 20, BufferItems: 64})
	defer kv.Close()
	if kv.BackendName() != "memory" {
		t.Fatalf("BackendName: got %q, want %q", kv.BackendName(), "memory")
	}
}

// --- SQLite backend tests ---

func TestSQLiteBackend_RoundTrip(t *testing.T) {
	kv, err := newSQLiteBackend(Options{SQLitePath: ":memory:", MaxValueLen: 10240, Maintenance: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	if err := kv.Set(ctx, "hello", []byte("world"), time.Hour); err != nil {
		t.Fatal(err)
	}
	got, ok, err := kv.Get(ctx, "hello")
	if err != nil || !ok {
		t.Fatalf("Get: err=%v ok=%v", err, ok)
	}
	if string(got) != "world" {
		t.Fatalf("got %q, want %q", got, "world")
	}
}

func TestSQLiteBackend_TTLExpiry(t *testing.T) {
	kv, err := newSQLiteBackend(Options{SQLitePath: ":memory:", MaxValueLen: 10240, Maintenance: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	kv.Set(ctx, "temp", []byte("x"), 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	_, ok, err := kv.Get(ctx, "temp")
	if err != nil || ok {
		t.Fatalf("expected expiry: ok=%v, err=%v", ok, err)
	}
}

func TestSQLiteBackend_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	kv, err := newSQLiteBackend(Options{SQLitePath: path, MaxValueLen: 10240, Maintenance: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	kv.Set(context.Background(), "persist", []byte("forever"), 0) // no TTL
	kv.Close()

	// Reopen
	kv2, err := newSQLiteBackend(Options{SQLitePath: path, MaxValueLen: 10240, Maintenance: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	defer kv2.Close()
	got, ok, err := kv2.Get(context.Background(), "persist")
	if err != nil || !ok || string(got) != "forever" {
		t.Fatalf("persistence failed: ok=%v, err=%v, got=%q", ok, err, got)
	}
}

func TestSQLiteBackend_IncrConcurrency(t *testing.T) {
	kv, err := newSQLiteBackend(Options{SQLitePath: ":memory:", MaxValueLen: 10240, Maintenance: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	n := 50
	incPerGoroutine := 100
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incPerGoroutine; j++ {
				if _, err := kv.Incr(ctx, "sqlite_counter", time.Hour); err != nil {
					errs <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Fatal(e)
	}

	v, err := kv.Incr(ctx, "sqlite_counter", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	expected := int64(n*incPerGoroutine + 1)
	if v != expected {
		t.Fatalf("concurrent SQLite Incr: got %d, want %d", v, expected)
	}
}

func TestSQLiteBackend_BackendName(t *testing.T) {
	kv, _ := newSQLiteBackend(Options{SQLitePath: ":memory:"})
	defer kv.Close()
	if kv.BackendName() != "sqlite" {
		t.Fatalf("BackendName: got %q, want %q", kv.BackendName(), "sqlite")
	}
}

func TestMemoryBackend_ValueTooLarge(t *testing.T) {
	kv, _ := newMemoryBackend(Options{NumCounters: 100, MaxCost: 1 << 20, BufferItems: 64, MaxValueLen: 10})
	defer kv.Close()
	err := kv.Set(context.Background(), "k", make([]byte, 100), time.Hour)
	if err == nil {
		t.Fatal("expected error for oversized value")
	}
}

// --- Redis/Valkey backend tests ---

func TestRedisBackend_RoundTrip(t *testing.T) {
	redisAddr := os.Getenv("SEARGO_TEST_REDIS_ADDR")
	if redisAddr == "" {
		t.Skip("SEARGO_TEST_REDIS_ADDR not set; skipping Redis integration test")
	}
	kv, err := newRedisBackend(Options{ValkeyURL: redisAddr, MaxValueLen: 10240})
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v", err)
	}
	defer kv.Close()

	ctx := context.Background()
	key := "test:roundtrip:" + uuid.New().String()
	if err := kv.Set(ctx, key, []byte("valkey_value"), time.Minute); err != nil {
		t.Fatal(err)
	}
	got, ok, err := kv.Get(ctx, key)
	if err != nil || !ok || string(got) != "valkey_value" {
		t.Fatalf("Get: err=%v ok=%v got=%q", err, ok, got)
	}
	kv.Delete(ctx, key)
}

func TestRedisBackend_Incr(t *testing.T) {
	redisAddr := os.Getenv("SEARGO_TEST_REDIS_ADDR")
	if redisAddr == "" {
		t.Skip("SEARGO_TEST_REDIS_ADDR not set")
	}
	kv, err := newRedisBackend(Options{ValkeyURL: redisAddr})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	key := "test:incr:" + uuid.New().String()
	defer kv.Delete(ctx, key)

	v, err := kv.Incr(ctx, key, time.Minute)
	if err != nil || v != 1 {
		t.Fatalf("Incr #1: got %d err=%v, want 1", v, err)
	}
	v, err = kv.Incr(ctx, key, time.Minute)
	if err != nil || v != 2 {
		t.Fatalf("Incr #2: got %d err=%v, want 2", v, err)
	}
}

func TestRedisBackend_BackendName(t *testing.T) {
	kv := &redisBackend{client: nil, backendName: "valkey"}
	if kv.BackendName() != "valkey" {
		t.Fatalf("BackendName: got %q, want %q", kv.BackendName(), "valkey")
	}
}

// --- Namespaced KV tests ---

func TestNamespacedKV_KeyIsolation(t *testing.T) {
	kv, err := New(Options{Backend: "memory", NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	ns1 := kv.WithNamespace("app1")
	ns2 := kv.WithNamespace("app2")

	ns1.Set(ctx, "key", []byte("value1"), time.Hour)
	ns2.Set(ctx, "key", []byte("value2"), time.Hour)

	got1, ok, _ := ns1.Get(ctx, "key")
	if !ok || string(got1) != "value1" {
		t.Fatalf("ns1: got %q, ok=%v", got1, ok)
	}
	got2, ok, _ := ns2.Get(ctx, "key")
	if !ok || string(got2) != "value2" {
		t.Fatalf("ns2: got %q, ok=%v", got2, ok)
	}
}

func TestNamespacedKV_Chained(t *testing.T) {
	kv, err := New(Options{Backend: "memory", NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	ns := kv.WithNamespace("a").WithNamespace("b")

	ns.Set(ctx, "k", []byte("v"), time.Hour)
	got, ok, _ := ns.Get(ctx, "k")
	if !ok || string(got) != "v" {
		t.Fatalf("chained ns: got %q, ok=%v", got, ok)
	}
}

func TestNamespacedKV_BackendName(t *testing.T) {
	kv, _ := New(Options{Backend: "memory", NumCounters: 100, MaxCost: 1 << 20, BufferItems: 64})
	defer kv.Close()
	ns := kv.WithNamespace("test")
	if ns.BackendName() != "memory" {
		t.Fatalf("BackendName: got %q, want %q", ns.BackendName(), "memory")
	}
}

// --- Hashed KV tests ---

func TestHashedKV_RoundTrip(t *testing.T) {
	raw, err := newMemoryBackend(Options{NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()

	mac := hmac.New(sha256.New, []byte("secret"))
	kv := &hashedKV{parent: raw, mac: mac}

	ctx := context.Background()
	kv.Set(ctx, "mykey", []byte("myvalue"), time.Hour)
	got, ok, _ := kv.Get(ctx, "mykey")
	if !ok || string(got) != "myvalue" {
		t.Fatalf("hashed: got %q, ok=%v", got, ok)
	}
}

func TestHashedKV_DifferentSecret(t *testing.T) {
	raw, err := newMemoryBackend(Options{NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()

	mac1 := hmac.New(sha256.New, []byte("secret1"))
	kv1 := &hashedKV{parent: raw, mac: mac1}

	mac2 := hmac.New(sha256.New, []byte("secret2"))
	kv2 := &hashedKV{parent: raw, mac: mac2}

	ctx := context.Background()
	kv1.Set(ctx, "samekey", []byte("v1"), time.Hour)
	got, ok, _ := kv2.Get(ctx, "samekey")
	if ok {
		t.Fatal("expected miss with different hash secret")
	}
	_ = got
}

// --- Factory tests ---

func TestNew_DefaultMemory(t *testing.T) {
	kv, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()
	if kv.BackendName() != "memory" {
		t.Fatalf("BackendName: got %q, want %q", kv.BackendName(), "memory")
	}

	ctx := context.Background()
	kv.Set(ctx, "k", []byte("v"), time.Hour)
	got, ok, _ := kv.Get(ctx, "k")
	if !ok || string(got) != "v" {
		t.Fatalf("got %q, ok=%v", got, ok)
	}
}

func TestNew_UnknownBackend(t *testing.T) {
	_, err := New(Options{Backend: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
}

func TestNew_SQLiteBackend(t *testing.T) {
	kv, err := New(Options{Backend: "sqlite", SQLitePath: ":memory:", MaxValueLen: 10240, Maintenance: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	kv.Set(ctx, "k", []byte("v"), time.Hour)
	got, ok, _ := kv.Get(ctx, "k")
	if !ok || string(got) != "v" {
		t.Fatalf("got %q, ok=%v", got, ok)
	}
}

func TestNew_WithKeyHash(t *testing.T) {
	kv, err := New(Options{
		Backend:       "memory",
		KeyHashSecret: "mysecret",
		NumCounters:   1000,
		MaxCost:       1 << 20,
		BufferItems:   64,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ctx := context.Background()
	kv.Set(ctx, "k", []byte("v"), time.Hour)
	got, ok, _ := kv.Get(ctx, "k")
	if !ok || string(got) != "v" {
		t.Fatalf("got %q, ok=%v", got, ok)
	}
}

func TestNew_WithNamespace(t *testing.T) {
	kv, err := New(Options{Backend: "memory", NumCounters: 1000, MaxCost: 1 << 20, BufferItems: 64})
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()

	ns := kv.WithNamespace("search")
	_, ok := ns.(*namespacedKV)
	if !ok {
		t.Fatal("WithNamespace should return a namespacedKV")
	}
}
