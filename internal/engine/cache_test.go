package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineCache_SetGet(t *testing.T) {
	cache, err := NewEngineCache(":memory:")
	require.NoError(t, err)
	defer cache.Close()

	err = cache.Set("test_engine", "key1", "value1", 60)
	require.NoError(t, err)

	val, ok := cache.Get("test_engine", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)
}

func TestEngineCache_Expired(t *testing.T) {
	cache, err := NewEngineCache(":memory:")
	require.NoError(t, err)
	defer cache.Close()

	err = cache.Set("test_engine", "key2", "value2", -1)
	require.NoError(t, err)

	val, ok := cache.Get("test_engine", "key2")
	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestEngineCache_MissingKey(t *testing.T) {
	cache, err := NewEngineCache(":memory:")
	require.NoError(t, err)
	defer cache.Close()

	_, ok := cache.Get("test_engine", "nonexistent")
	assert.False(t, ok)
}

func TestEngineCache_Overwrite(t *testing.T) {
	cache, err := NewEngineCache(":memory:")
	require.NoError(t, err)
	defer cache.Close()

	cache.Set("eng", "k", "v1", 60)
	cache.Set("eng", "k", "v2", 60)

	val, ok := cache.Get("eng", "k")
	assert.True(t, ok)
	assert.Equal(t, "v2", val)
}

func TestEngineCache_DifferentEngines(t *testing.T) {
	cache, err := NewEngineCache(":memory:")
	require.NoError(t, err)
	defer cache.Close()

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
	cache, err := NewEngineCache(":memory:")
	require.NoError(t, err)
	defer cache.Close()

	cache.Set("eng", "k", "v", 60)
	cache.Delete("eng", "k")

	_, ok := cache.Get("eng", "k")
	assert.False(t, ok)
}

func TestEngineCache_FilePersistence(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/cache.db"

	cache, err := NewEngineCache(path)
	require.NoError(t, err)

	cache.Set("eng", "k", "v", 3600)
	cache.Close()

	// Reopen
	cache2, err := NewEngineCache(path)
	require.NoError(t, err)
	defer cache2.Close()

	val, ok := cache2.Get("eng", "k")
	assert.True(t, ok)
	assert.Equal(t, "v", val)
}
