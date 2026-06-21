package autocomplete

import (
	"context"
	"encoding/json"
	"time"

	"github.com/seargo/seargo/internal/storage"
)

const DefaultCacheTTL = 45 * time.Second

// ResultCache provides a TTL-based cache for autocomplete suggestions, backed by storage.KV.
type ResultCache struct {
	kv  storage.KV
	ttl time.Duration
}

// NewResultCache creates a ResultCache backed by a KV store.
func NewResultCache(kv storage.KV, ttl time.Duration) *ResultCache {
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	return &ResultCache{kv: kv, ttl: ttl}
}

// Get retrieves cached results for a key.
func (c *ResultCache) Get(key string) ([]string, bool) {
	raw, ok, err := c.kv.Get(context.Background(), key)
	if err != nil || !ok {
		return nil, false
	}
	var results []string
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, false
	}
	return results, true
}

// Set stores results for a key with the configured TTL.
func (c *ResultCache) Set(key string, results []string) {
	raw, err := json.Marshal(results)
	if err != nil {
		return
	}
	_ = c.kv.Set(context.Background(), key, raw, c.ttl)
}

// Close is a no-op. The shared KV is owned by the caller.
func (c *ResultCache) Close() {}
