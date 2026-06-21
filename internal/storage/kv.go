package storage

import (
	"context"
	"strings"
	"time"
)

// KV is the shared key-value storage interface used by all stateful consumers.
type KV interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Close() error
	WithNamespace(namespace string) KV
	BackendName() string
}

// Options configures a KV backend instance.
type Options struct {
	Backend       string
	ValkeyURL     string
	SQLitePath    string
	MaxValueLen   int
	KeyHashSecret string
	Maintenance   time.Duration
	NumCounters   int64
	MaxCost       int64
	BufferItems   int64
}

// sanitize replaces characters not in [a-zA-Z0-9_:-] with '_'.
func sanitize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == ':' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
