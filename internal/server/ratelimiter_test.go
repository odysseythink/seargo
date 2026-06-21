package server

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, 10*time.Second) // 3 tokens per 10 seconds
	defer rl.Close()

	ip := "10.0.0.1"
	for i := 0; i < 3; i++ {
		if !rl.Allow(ip) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if rl.Allow(ip) {
		t.Fatal("4th request should be denied")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)
	defer rl.Close()

	ip := "10.0.0.2"
	rl.Allow(ip) // consume 1
	rl.Allow(ip) // consume 1
	if rl.Allow(ip) {
		t.Fatal("3rd request should be denied before refill")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow(ip) {
		t.Fatal("should be allowed after refill interval")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	defer rl.Close()

	rl.Allow("10.0.0.1") // exhaust ip1
	if !rl.Allow("10.0.0.2") {
		t.Fatal("different IP should have its own bucket")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(3, 10*time.Millisecond)
	defer rl.Close()

	rl.Allow("10.0.0.3")
	time.Sleep(50 * time.Millisecond)
	rl.cleanup()

	rl.mu.Lock()
	_, exists := rl.buckets["10.0.0.3"]
	rl.mu.Unlock()
	if exists {
		t.Fatal("stale bucket should be cleaned up")
	}
}

func TestRateLimiter_DefaultRate(t *testing.T) {
	rl := NewRateLimiter(0, 0) // use defaults
	defer rl.Close()
	if rl.rate != 30 {
		t.Errorf("expected default rate 30, got %d", rl.rate)
	}
}
