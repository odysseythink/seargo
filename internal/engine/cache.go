package engine

import (
	"context"
	"time"

	"github.com/seargo/seargo/internal/storage"
)

// EngineCache provides a per-engine key/value store backed by storage.KV.
type EngineCache struct {
	base storage.KV
}

// NewEngineCache creates an EngineCache backed by a namespaced KV.
func NewEngineCache(kv storage.KV) *EngineCache {
	return &EngineCache{base: kv}
}

func (c *EngineCache) namespace(engineName string) storage.KV {
	return c.base.WithNamespace(engineName)
}

// Set stores a value with a TTL in seconds. ttl <= 0 means immediately expired (not stored).
func (c *EngineCache) Set(engineName, key, value string, ttl int64) error {
	// ttl <= 0 means the old engine cache treated this as "immediately expired"
	if ttl <= 0 {
		return nil
	}
	kv := c.namespace(engineName)
	d := time.Duration(ttl) * time.Second
	return kv.Set(context.Background(), key, []byte(value), d)
}

// Get retrieves a value. Returns (value, true) if found and not expired.
func (c *EngineCache) Get(engineName, key string) (string, bool) {
	kv := c.namespace(engineName)
	raw, ok, err := kv.Get(context.Background(), key)
	if err != nil || !ok {
		return "", false
	}
	return string(raw), true
}

// Delete removes a key for an engine.
func (c *EngineCache) Delete(engineName, key string) error {
	return c.namespace(engineName).Delete(context.Background(), key)
}

// Close is a no-op (shared KV is owned by caller).
func (c *EngineCache) Close() error {
	return nil
}

// PurgeExpired is a no-op (handled by storage backend).
func (c *EngineCache) PurgeExpired() error {
	return nil
}
