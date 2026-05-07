package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"

	"github.com/seargo/seargo/pkg/models"
)

type MultiLevel struct {
	local            *ristretto.Cache
	remote           *redis.Client
	defaultLocalTTL  time.Duration
	defaultRemoteTTL time.Duration
}

func NewMultiLevel(redisAddr string) (*MultiLevel, error) {
	localCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 28, // 256MB
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("create local cache: %w", err)
	}

	var rdb *redis.Client
	if redisAddr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
	}

	return &MultiLevel{
		local:            localCache,
		remote:           rdb,
		defaultLocalTTL:  30 * time.Second,
		defaultRemoteTTL: 5 * time.Minute,
	}, nil
}

func (m *MultiLevel) Get(key string) (*models.Response, bool) {
	// L1: local cache
	if val, ok := m.local.Get(key); ok {
		if resp, ok := val.(*models.Response); ok {
			return resp, true
		}
	}

	// L2: Redis
	if m.remote != nil {
		val, err := m.remote.Get(context.Background(), key).Result()
		if err == nil {
			var resp models.Response
			if err := json.Unmarshal([]byte(val), &resp); err == nil {
				m.local.SetWithTTL(key, &resp, 1, m.defaultLocalTTL)
				return &resp, true
			}
		}
	}

	return nil, false
}

func (m *MultiLevel) Set(key string, value *models.Response, ttl time.Duration) {
	m.local.SetWithTTL(key, value, 1, ttl)
	m.local.Wait()

	if m.remote != nil {
		if data, err := json.Marshal(value); err == nil {
			m.remote.Set(context.Background(), key, data, ttl)
		}
	}
}

func (m *MultiLevel) Delete(key string) {
	m.local.Del(key)
	if m.remote != nil {
		m.remote.Del(context.Background(), key)
	}
}
