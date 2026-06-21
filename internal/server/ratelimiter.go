package server

import (
	"sync"
	"time"
)

// DefaultRateLimit is the default rate limit per IP: 30 requests per minute.
const DefaultRateLimit = 30

// ipBucket tracks remaining tokens and last refill time for a single IP.
type ipBucket struct {
	tokens    int
	lastRefill time.Time
}

// RateLimiter implements a per-IP token-bucket rate limiter with automatic
// cleanup of stale entries.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*ipBucket
	rate     int // tokens per refill interval
	interval time.Duration
	stop     chan struct{}
	closeOnce sync.Once
}

// NewRateLimiter creates a rate limiter. rate is tokens per interval.
// If rate <= 0, DefaultRateLimit is used. If interval <= 0, 1 minute is used.
// Starts a background cleanup goroutine — call Close() when done.
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	if rate <= 0 {
		rate = DefaultRateLimit
	}
	if interval <= 0 {
		interval = time.Minute
	}
	rl := &RateLimiter{
		buckets:  make(map[string]*ipBucket),
		rate:     rate,
		interval: interval,
		stop:     make(chan struct{}),
	}
	go rl.cleanupLoop(10 * time.Minute)
	return rl
}

// Allow returns true if the request from ip should be allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok || now.Sub(b.lastRefill) >= rl.interval {
		// New IP or interval elapsed — reset bucket
		rl.buckets[ip] = &ipBucket{
			tokens:    rl.rate - 1, // consume one token
			lastRefill: now,
		}
		return true
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// Close stops the background cleanup goroutine. Safe to call multiple times.
func (rl *RateLimiter) Close() {
	rl.closeOnce.Do(func() {
		close(rl.stop)
	})
}

func (rl *RateLimiter) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	expiry := time.Now().Add(-rl.interval * 2)
	for ip, b := range rl.buckets {
		if b.lastRefill.Before(expiry) {
			delete(rl.buckets, ip)
		}
	}
}
