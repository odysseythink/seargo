package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/seargo/seargo/internal/storage"
)

func makeTestKV(t *testing.T) storage.KV {
	t.Helper()
	kv, err := storage.New(storage.Options{
		Backend:     "memory",
		NumCounters: 10000,
		MaxCost:     10 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { kv.Close() })
	return kv.WithNamespace("limiter_test")
}

func TestDropCounter(t *testing.T) {
	kv := makeTestKV(t)
	lm := New(&Config{
		BurstWindow:     time.Second,
		BurstMax:        5,
		LongWindow:      time.Second,
		LongMax:         10,
		FilterLinkLocal: false,
		LinkToken:       false,
	}, kv)

	key := "test:drop"
	lm.DropCounter(key)

	// Verify counter key is gone (should return 0 or empty)
	v, _, err := kv.Get(context.Background(), "counter:"+key)
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Fatal("expected counter to be dropped")
	}
}

func TestAllow_UnderLimit(t *testing.T) {
	kv := makeTestKV(t)
	lm := New(&Config{
		BurstWindow:     time.Minute,
		BurstMax:        10,
		LongWindow:      time.Minute,
		LongMax:         20,
		FilterLinkLocal: false,
		LinkToken:       false,
	}, kv)

	// 5 requests should be under both limits
	for i := 0; i < 5; i++ {
		allowed, _, err := lm.Allow(context.Background(), "1.2.3.4", false)
		if err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i)
		}
	}
}

func TestAllow_BurstLimit(t *testing.T) {
	kv := makeTestKV(t)
	lm := New(&Config{
		BurstWindow:     time.Minute,
		BurstMax:        2,
		LongWindow:      time.Minute,
		LongMax:         20,
		FilterLinkLocal: false,
		LinkToken:       false,
	}, kv)

	// 2 requests should be allowed
	for i := 0; i < 2; i++ {
		allowed, _, err := lm.Allow(context.Background(), "1.2.3.4", false)
		if err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	// 3rd should hit burst limit
	allowed, reason, err := lm.Allow(context.Background(), "1.2.3.4", false)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("3rd request should be blocked by burst limit")
	}
	if reason != "burst" {
		t.Fatalf("reason should be 'burst', got %q", reason)
	}
}

func TestAllow_LongLimit(t *testing.T) {
	kv := makeTestKV(t)
	lm := New(&Config{
		BurstWindow:     time.Minute,
		BurstMax:        10,
		LongWindow:      time.Minute,
		LongMax:         2,
		FilterLinkLocal: false,
		LinkToken:       false,
	}, kv)

	allowed, _, err := lm.Allow(context.Background(), "1.2.3.4", false)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("1st request should be allowed")
	}

	allowed, _, err = lm.Allow(context.Background(), "1.2.3.4", false)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("2nd request should be allowed")
	}

	allowed, reason, err := lm.Allow(context.Background(), "1.2.3.4", false)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("3rd request should be blocked")
	}
	if reason != "long" {
		t.Fatalf("reason should be 'long', got %q", reason)
	}
}

func TestAllow_LinkLocalExempt(t *testing.T) {
	kv := makeTestKV(t)
	lm := New(&Config{
		BurstWindow:     time.Minute,
		BurstMax:        0,
		LongWindow:      time.Minute,
		LongMax:         0,
		FilterLinkLocal: true,
		LinkToken:       false,
	}, kv)

	// Link-local IPs should be exempt
	allowed, _, err := lm.Allow(context.Background(), "169.254.1.1", false)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("link-local IP should be exempt from rate limiting")
	}
}

func TestAllow_APIRateLimit(t *testing.T) {
	kv := makeTestKV(t)
	lm := New(&Config{
		BurstWindow:     time.Minute,
		BurstMax:        10,
		LongWindow:      time.Minute,
		LongMax:         20,
		FilterLinkLocal: false,
		LinkToken:       false,
		APIMax:          2,
		APIWindow:       time.Minute,
	}, kv)

	// First 2 API requests should be allowed
	for i := 0; i < 2; i++ {
		allowed, _, err := lm.Allow(context.Background(), "1.2.3.4", true)
		if err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Fatalf("API request %d should be allowed", i)
		}
	}

	// 3rd API request should be blocked
	allowed, reason, err := lm.Allow(context.Background(), "1.2.3.4", true)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("3rd API request should be blocked")
	}
	if reason != "api" {
		t.Fatalf("reason should be 'api', got %q", reason)
	}
}

func TestLoadLimiterConfig(t *testing.T) {
	cfg, err := LoadConfig("../configs/limiter.toml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("cfg is nil")
	}
	if cfg.BurstMax != 15 {
		t.Fatalf("BurstMax: got %d, want 15", cfg.BurstMax)
	}
}
