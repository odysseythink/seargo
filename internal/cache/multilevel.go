package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/metrics"
	"github.com/seargo/seargo/internal/storage"
	"github.com/seargo/seargo/pkg/models"
)

// Config holds search-result cache settings.
type Config struct {
	Enabled       bool
	LocalTTL      int
	RemoteTTL     int
	TTLByCategory map[models.Category]int
}

// MultiLevel is a two-level search result cache backed by storage.KV.
type MultiLevel struct {
	enabled   bool
	l1        storage.KV // local (memory)
	l2        storage.KV // shared (configured backend)
	localTTL  time.Duration
	remoteTTL time.Duration
	ttlByCat  map[models.Category]time.Duration
	namespace string
}

// NewMultiLevel creates a MultiLevel cache.
// shared is the backend selected by storage.NewFromConfig; l1 is always a memory backend.
func NewMultiLevel(shared storage.KV, cfg Config) (*MultiLevel, error) {
	l1, err := storage.New(storage.Options{
		Backend:     "memory",
		NumCounters: 10_000_000,
		MaxCost:     256 << 20,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	if cfg.LocalTTL <= 0 {
		cfg.LocalTTL = 30
	}
	if cfg.RemoteTTL <= 0 {
		cfg.RemoteTTL = 300
	}

	ttlByCat := make(map[models.Category]time.Duration, len(cfg.TTLByCategory))
	for cat, secs := range cfg.TTLByCategory {
		if secs > 0 {
			ttlByCat[cat] = time.Duration(secs) * time.Second
		}
	}

	return &MultiLevel{
		enabled:   cfg.Enabled,
		l1:        l1,
		l2:        shared.WithNamespace("search"),
		localTTL:  time.Duration(cfg.LocalTTL) * time.Second,
		remoteTTL: time.Duration(cfg.RemoteTTL) * time.Second,
		ttlByCat:  ttlByCat,
		namespace: "search",
	}, nil
}

func (m *MultiLevel) storageKey(key string) string {
	return key
}

func (m *MultiLevel) ttlForCategory(cat models.Category) time.Duration {
	if d, ok := m.ttlByCat[cat]; ok && d > 0 {
		return d
	}
	return m.localTTL
}

// Get implements the Cache interface.
func (m *MultiLevel) Get(key string) (*models.Response, bool) {
	if !m.enabled {
		return nil, false
	}
	ctx := context.Background()

	// L1: local
	if raw, ok, err := m.l1.Get(ctx, m.storageKey(key)); err == nil && ok {
		resp, err := unmarshalResponse(raw)
		if err == nil {
			metrics.CacheHits.WithLabelValues("local").Inc()
			return resp, true
		}
	}

	// L2: shared
	raw, ok, err := m.l2.Get(ctx, m.storageKey(key))
	if err != nil || !ok {
		metrics.CacheMisses.WithLabelValues("all").Inc()
		return nil, false
	}

	resp, err := unmarshalResponse(raw)
	if err != nil {
		metrics.CacheMisses.WithLabelValues("all").Inc()
		return nil, false
	}

	// Promote to L1
	_ = m.l1.Set(ctx, m.storageKey(key), raw, m.ttlForCategory(resp.Category))
	metrics.CacheHits.WithLabelValues("remote").Inc()
	return resp, true
}

// Set implements the Cache interface.
// ttl <= 0 means use the category-based TTL.
func (m *MultiLevel) Set(key string, value *models.Response, ttl time.Duration) {
	if !m.enabled {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		logger.Warn("cache: marshal failed", "key", key, "error", err)
		return
	}
	if ttl <= 0 {
		ttl = m.ttlForCategory(value.Category)
	}

	ctx := context.Background()
	if err := m.l1.Set(ctx, m.storageKey(key), raw, ttl); err != nil {
		logger.Warn("cache: L1 write failed", "key", key, "error", err)
	}
	if err := m.l2.Set(ctx, m.storageKey(key), raw, ttl); err != nil {
		logger.Warn("cache: L2 write failed", "key", key, "error", err)
	}
}

// Delete implements the Cache interface.
func (m *MultiLevel) Delete(key string) {
	ctx := context.Background()
	_ = m.l1.Delete(ctx, m.storageKey(key))
	_ = m.l2.Delete(ctx, m.storageKey(key))
}

func unmarshalResponse(b []byte) (*models.Response, error) {
	var r models.Response
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
