package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"
)

// namespacedKV wraps a KV and prefixes all keys with a sanitized namespace.
type namespacedKV struct {
	parent    KV
	namespace string
}

func (n *namespacedKV) key(k string) string {
	return n.namespace + ":" + k
}

func (n *namespacedKV) Get(ctx context.Context, key string) ([]byte, bool, error) {
	return n.parent.Get(ctx, n.key(key))
}

func (n *namespacedKV) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return n.parent.Set(ctx, n.key(key), value, ttl)
}

func (n *namespacedKV) Delete(ctx context.Context, key string) error {
	return n.parent.Delete(ctx, n.key(key))
}

func (n *namespacedKV) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return n.parent.SetNX(ctx, n.key(key), value, ttl)
}

func (n *namespacedKV) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return n.parent.Incr(ctx, n.key(key), ttl)
}

func (n *namespacedKV) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return n.parent.Expire(ctx, n.key(key), ttl)
}

func (n *namespacedKV) Close() error {
	return n.parent.Close()
}

func (n *namespacedKV) WithNamespace(namespace string) KV {
	return &namespacedKV{parent: n, namespace: sanitize(namespace)}
}

func (n *namespacedKV) BackendName() string {
	return n.parent.BackendName()
}

// hashedKV wraps a KV and hashes keys with HMAC-SHA256 before passing to parent.
type hashedKV struct {
	parent KV
	mac    hash.Hash
}

func (h *hashedKV) hashKey(key string) string {
	h.mac.Reset()
	h.mac.Write([]byte(key))
	return hex.EncodeToString(h.mac.Sum(nil))
}

func (h *hashedKV) Get(ctx context.Context, key string) ([]byte, bool, error) {
	return h.parent.Get(ctx, h.hashKey(key))
}

func (h *hashedKV) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return h.parent.Set(ctx, h.hashKey(key), value, ttl)
}

func (h *hashedKV) Delete(ctx context.Context, key string) error {
	return h.parent.Delete(ctx, h.hashKey(key))
}

func (h *hashedKV) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return h.parent.SetNX(ctx, h.hashKey(key), value, ttl)
}

func (h *hashedKV) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return h.parent.Incr(ctx, h.hashKey(key), ttl)
}

func (h *hashedKV) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return h.parent.Expire(ctx, h.hashKey(key), ttl)
}

func (h *hashedKV) Close() error {
	return h.parent.Close()
}

func (h *hashedKV) WithNamespace(namespace string) KV {
	// Wrap in namespacedKV so namespace prefix is applied before hashing.
	return &namespacedKV{parent: h, namespace: sanitize(namespace)}
}

func (h *hashedKV) BackendName() string {
	return h.parent.BackendName()
}

// New creates a KV backend from the given Options.
func New(opts Options) (KV, error) {
	var raw KV
	var err error

	switch opts.Backend {
	case "", "memory":
		raw, err = newMemoryBackend(opts)
	case "sqlite":
		raw, err = newSQLiteBackend(opts)
	case "valkey":
		raw, err = newRedisBackend(opts)
	default:
		return nil, fmt.Errorf("unknown storage backend: %q", opts.Backend)
	}
	if err != nil {
		return nil, err
	}

	// Wrap in namespacedKV so WithNamespace works on the returned KV.
	var kv KV = &namespacedKV{parent: raw, namespace: ""}

	// Wrap with HMAC key hashing if a secret is configured.
	if opts.KeyHashSecret != "" {
		mac := hmac.New(sha256.New, []byte(opts.KeyHashSecret))
		kv = &hashedKV{parent: kv, mac: mac}
	}

	return kv, nil
}

// NewFromConfig creates a KV backend from the application configuration.
func NewFromConfig(cfg interface {
	StorageBackend() string
	StorageValkeyURL() string
	StorageSQLitePath() string
	StorageMaxValueLen() int
	StorageKeyHashSecret() string
	StorageMaintenance() time.Duration
	StorageNumCounters() int64
	StorageMaxCost() int64
	StorageBufferItems() int64
}) (KV, error) {
	return New(Options{
		Backend:       cfg.StorageBackend(),
		ValkeyURL:     cfg.StorageValkeyURL(),
		SQLitePath:    cfg.StorageSQLitePath(),
		MaxValueLen:   cfg.StorageMaxValueLen(),
		KeyHashSecret: cfg.StorageKeyHashSecret(),
		Maintenance:   cfg.StorageMaintenance(),
		NumCounters:   cfg.StorageNumCounters(),
		MaxCost:       cfg.StorageMaxCost(),
		BufferItems:   cfg.StorageBufferItems(),
	})
}

// compile-time interface checks
var _ KV = (*namespacedKV)(nil)
var _ KV = (*hashedKV)(nil)
