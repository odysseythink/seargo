package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/seargo/seargo/internal/storage"
)

// counter manages fixed-window counters using storage.KV.
type counter struct {
	kv storage.KV
}

func newCounter(kv storage.KV) *counter {
	return &counter{kv: kv}
}

// Incr increments the counter for the given key and window.
// The window duration is used as the TTL — when it expires the counter resets.
func (c *counter) Incr(ctx context.Context, key string, window time.Duration, max int64) (int64, bool, error) {
	counterKey := fmt.Sprintf("counter:%s", key)
	val, err := c.kv.Incr(ctx, counterKey, window)
	if err != nil {
		return 0, false, err
	}
	return val, val <= max, nil
}

// Drop removes the counter for the given key.
func (c *counter) Drop(ctx context.Context, key string) {
	c.kv.Delete(ctx, fmt.Sprintf("counter:%s", key))
}
