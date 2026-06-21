package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisBackend struct {
	client      *redis.Client
	maxValueLen int
	backendName string
}

func newRedisBackend(opts Options) (*redisBackend, error) {
	url := opts.ValkeyURL
	if url == "" {
		return nil, fmt.Errorf("redis backend: valkey URL is required")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis backend: parse URL: %w", err)
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis backend: ping: %w", err)
	}
	return &redisBackend{
		client:      client,
		maxValueLen: opts.MaxValueLen,
		backendName: "valkey",
	}, nil
}

func (r *redisBackend) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return []byte(val), true, nil
}

func (r *redisBackend) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if r.maxValueLen > 0 && len(value) > r.maxValueLen {
		return fmt.Errorf("value size %d exceeds max %d", len(value), r.maxValueLen)
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisBackend) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisBackend) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (r *redisBackend) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl > 0 {
		if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, err
		}
	}
	return val, nil
}

func (r *redisBackend) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

func (r *redisBackend) Close() error {
	return r.client.Close()
}

func (r *redisBackend) WithNamespace(namespace string) KV {
	panic("not implemented: use through New()")
}

func (r *redisBackend) BackendName() string {
	return r.backendName
}
