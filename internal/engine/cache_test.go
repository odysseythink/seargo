package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/storage"
)

func makeTestKV(t *testing.T) storage.KV {
	t.Helper()
	kv, err := storage.New(storage.Options{
		Backend:     "memory",
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	require.NoError(t, err)
	t.Cleanup(func() { kv.Close() })
	return kv
}

func TestEngineCache_SetGet(t *testing.T) {
	kv := makeTestKV(t)
	cache := NewEngineCache(kv.WithNamespace("engine"))

	err := cache.Set("test_engine", "key1", "value1", 60)
	require.NoError(t, err)

	val, ok := cache.Get("test_engine", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)
}

func TestEngineCache_Expired(t *testing.T) {
	kv := makeTestKV(t)
	cache := NewEngineCache(kv.WithNamespace("engine"))

	err := cache.Set("test_engine", "key2", "value2", -1)
	require.NoError(t, err)

	val, ok := cache.Get("test_engine", "key2")
	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestEngineCache_MissingKey(t *testing.T) {
	kv := makeTestKV(t)
	cache := NewEngineCache(kv.WithNamespace("engine"))

	_, ok := cache.Get("test_engine", "nonexistent")
	assert.False(t, ok)
}

func TestEngineCache_Overwrite(t *testing.T) {
	kv := makeTestKV(t)
	cache := NewEngineCache(kv.WithNamespace("engine"))

	cache.Set("eng", "k", "v1", 60)
	cache.Set("eng", "k", "v2", 60)

	val, ok := cache.Get("eng", "k")
	assert.True(t, ok)
	assert.Equal(t, "v2", val)
}

func TestEngineCache_DifferentEngines(t *testing.T) {
	kv := makeTestKV(t)
	cache := NewEngineCache(kv.WithNamespace("engine"))

	cache.Set("eng1", "k", "v1", 60)
	cache.Set("eng2", "k", "v2", 60)

	v1, ok := cache.Get("eng1", "k")
	assert.True(t, ok)
	assert.Equal(t, "v1", v1)

	v2, ok := cache.Get("eng2", "k")
	assert.True(t, ok)
	assert.Equal(t, "v2", v2)
}

func TestEngineCache_Delete(t *testing.T) {
	kv := makeTestKV(t)
	cache := NewEngineCache(kv.WithNamespace("engine"))

	cache.Set("eng", "k", "v", 60)
	cache.Delete("eng", "k")

	_, ok := cache.Get("eng", "k")
	assert.False(t, ok)
}
