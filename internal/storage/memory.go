package storage

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
)

type memoryBackend struct {
	cache       *ristretto.Cache
	maxValueLen int
	locks       *keyMutex
}

type keyMutex struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func (km *keyMutex) Lock(key string) {
	km.mu.Lock()
	m, ok := km.locks[key]
	if !ok {
		m = &sync.Mutex{}
		km.locks[key] = m
	}
	km.mu.Unlock()
	m.Lock()
}

func (km *keyMutex) Unlock(key string) {
	km.mu.Lock()
	m := km.locks[key]
	km.mu.Unlock()
	m.Unlock()
}

func newMemoryBackend(opts Options) (*memoryBackend, error) {
	nc := opts.NumCounters
	if nc <= 0 {
		nc = 10_000_000
	}
	mc := opts.MaxCost
	if mc <= 0 {
		mc = 256 << 20
	}
	bi := opts.BufferItems
	if bi <= 0 {
		bi = 64
	}
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: nc,
		MaxCost:     mc,
		BufferItems: bi,
	})
	if err != nil {
		return nil, fmt.Errorf("memory backend: %w", err)
	}
	return &memoryBackend{
		cache:       cache,
		maxValueLen: opts.MaxValueLen,
		locks:       &keyMutex{locks: make(map[string]*sync.Mutex)},
	}, nil
}

func (m *memoryBackend) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, ok := m.cache.Get(key)
	if !ok {
		return nil, false, nil
	}
	bytes, ok := val.([]byte)
	if !ok {
		return nil, false, fmt.Errorf("corrupt memory cache value for key %q", key)
	}
	return bytes, true, nil
}

func (m *memoryBackend) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.maxValueLen > 0 && len(value) > m.maxValueLen {
		return fmt.Errorf("value size %d exceeds max %d", len(value), m.maxValueLen)
	}
	cost := int64(len(value))
	if cost < 1 {
		cost = 1
	}
	ok := m.cache.SetWithTTL(key, value, cost, ttl)
	m.cache.Wait()
	if !ok {
		// ristretto can silently drop items when overloaded; treat as success
		// (best-effort cache semantics)
		return nil
	}
	return nil
}

func (m *memoryBackend) Delete(ctx context.Context, key string) error {
	m.cache.Del(key)
	return nil
}

func (m *memoryBackend) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	m.locks.Lock(key)
	defer m.locks.Unlock(key)

	_, ok := m.cache.Get(key)
	if ok {
		return false, nil
	}
	_ = m.Set(ctx, key, value, ttl)
	return true, nil
}

func (m *memoryBackend) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	m.locks.Lock(key)
	defer m.locks.Unlock(key)

	raw, found := m.cache.Get(key)
	if !found {
		m.cache.SetWithTTL(key, []byte("1"), 1, ttl)
		m.cache.Wait()
		return 1, nil
	}
	bytes, ok := raw.([]byte)
	if !ok {
		return 0, fmt.Errorf("corrupt counter value for key %q", key)
	}
	old, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		old = 0
	}
	newVal := old + 1
	newBytes := []byte(strconv.FormatInt(newVal, 10))
	m.cache.SetWithTTL(key, newBytes, 1, ttl)
	m.cache.Wait()
	return newVal, nil
}

func (m *memoryBackend) Expire(ctx context.Context, key string, ttl time.Duration) error {
	raw, ok := m.cache.Get(key)
	if !ok {
		return nil // no-op for missing key
	}
	bytes, ok := raw.([]byte)
	if !ok {
		return fmt.Errorf("corrupt value for key %q", key)
	}
	m.cache.SetWithTTL(key, bytes, int64(len(bytes)), ttl)
	m.cache.Wait()
	return nil
}

func (m *memoryBackend) Close() error {
	m.cache.Close()
	return nil
}

func (m *memoryBackend) WithNamespace(namespace string) KV {
	panic("not implemented: use memoryBackend through New() or wrap later")
}

func (m *memoryBackend) BackendName() string {
	return "memory"
}
