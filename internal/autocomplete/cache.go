package autocomplete

import (
	"sync"
	"time"
)

const DefaultCacheTTL = 45 * time.Second

type cacheEntry struct {
	results   []string
	expiresAt time.Time
}

type ResultCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
	stop  chan struct{}
	once  sync.Once
}

func NewResultCache(ttl time.Duration) *ResultCache {
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	c := &ResultCache{
		items: make(map[string]cacheEntry),
		ttl:   ttl,
		stop:  make(chan struct{}),
	}
	go c.cleanupLoop(5 * time.Minute)
	return c
}

func (c *ResultCache) Get(key string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.items[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.results, true
}

func (c *ResultCache) Set(key string, results []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheEntry{
		results:   results,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *ResultCache) Close() {
	c.once.Do(func() {
		close(c.stop)
	})
}

func (c *ResultCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.evictExpired()
		}
	}
}

func (c *ResultCache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.items {
		if now.After(v.expiresAt) {
			delete(c.items, k)
		}
	}
}
