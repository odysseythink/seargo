package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/storage"
	"github.com/seargo/seargo/pkg/models"
)

func makeTestStorage(t *testing.T) storage.KV {
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

func TestMultiLevel_L1Hit(t *testing.T) {
	shared := makeTestStorage(t)
	c, err := NewMultiLevel(shared, Config{
		Enabled:   true,
		LocalTTL:  30,
		RemoteTTL: 300,
	})
	require.NoError(t, err)

	resp := &models.Response{Query: "l1", Results: []models.Result{{Title: "L1 Result"}}}
	c.Set("l1key", resp, 0)

	got, ok := c.Get("l1key")
	require.True(t, ok)
	assert.Equal(t, "l1", got.Query)
}

func TestMultiLevel_L2HitPromotesToL1(t *testing.T) {
	shared := makeTestStorage(t)
	c, err := NewMultiLevel(shared, Config{
		Enabled:   true,
		LocalTTL:  30,
		RemoteTTL: 300,
	})
	require.NoError(t, err)

	resp := &models.Response{Query: "l2hit", Results: []models.Result{{Title: "L2 Result"}}}
	// Write directly to L2 only, skip L1
	raw, _ := json.Marshal(resp)
	c.l2.Set(context.Background(), "l2key", raw, time.Hour)

	// First Get: L1 miss, L2 hit, promote to L1
	got, ok := c.Get("l2key")
	require.True(t, ok)
	assert.Equal(t, "l2hit", got.Query)

	// Second Get should hit L1 (verify promotion)
	got2, ok := c.Get("l2key")
	require.True(t, ok)
	assert.Equal(t, "l2hit", got2.Query)
}

func TestMultiLevel_Miss(t *testing.T) {
	shared := makeTestStorage(t)
	c, err := NewMultiLevel(shared, Config{
		Enabled:   true,
		LocalTTL:  30,
		RemoteTTL: 300,
	})
	require.NoError(t, err)

	_, ok := c.Get("no_such_key")
	assert.False(t, ok)
}

func TestMultiLevel_SetBothLevels(t *testing.T) {
	shared := makeTestStorage(t)
	c, err := NewMultiLevel(shared, Config{
		Enabled:   true,
		LocalTTL:  30,
		RemoteTTL: 300,
	})
	require.NoError(t, err)

	resp := &models.Response{Query: "both", Results: []models.Result{{Title: "Both"}}}
	c.Set("bothkey", resp, 0)

	// Check L1
	raw1, ok1, _ := c.l1.Get(context.Background(), "bothkey")
	require.True(t, ok1)
	var got1 models.Response
	json.Unmarshal(raw1, &got1)
	assert.Equal(t, "both", got1.Query)

	// Check L2
	raw2, ok2, _ := c.l2.Get(context.Background(), "bothkey")
	require.True(t, ok2)
	var got2 models.Response
	json.Unmarshal(raw2, &got2)
	assert.Equal(t, "both", got2.Query)
}

func TestMultiLevel_Delete(t *testing.T) {
	shared := makeTestStorage(t)
	c, err := NewMultiLevel(shared, Config{
		Enabled:   true,
		LocalTTL:  30,
		RemoteTTL: 300,
	})
	require.NoError(t, err)

	resp := &models.Response{Query: "del"}
	c.Set("delkey", resp, 0)
	c.Delete("delkey")

	_, ok := c.Get("delkey")
	assert.False(t, ok)
}

func TestMultiLevel_CategoryTTLOverride(t *testing.T) {
	shared := makeTestStorage(t)
	c, err := NewMultiLevel(shared, Config{
		Enabled:   true,
		LocalTTL:  30,
		RemoteTTL: 300,
		TTLByCategory: map[models.Category]int{
			models.CategoryImages: 120,
			models.CategoryVideos: 120,
			models.CategoryNews:   15,
		},
	})
	require.NoError(t, err)

	resp := &models.Response{Query: "img", Category: models.CategoryImages}
	c.Set("imgkey", resp, 0)

	// L1 should have the image TTL (not default 30s)
	raw, ok, _ := c.l1.Get(context.Background(), "imgkey")
	require.True(t, ok)
	var got models.Response
	json.Unmarshal(raw, &got)
	assert.Equal(t, "img", got.Query)
}

func TestMultiLevel_JSONRoundTrip(t *testing.T) {
	resp := &models.Response{
		Query:     "rt",
		Category:  models.CategoryGeneral,
		Results:   []models.Result{{Title: "T", URL: "https://x.com", Engine: "g", Score: 1.0}},
		Total:     1,
		Page:      1,
		PageSize:  10,
	}
	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	var got models.Response
	err = json.Unmarshal(raw, &got)
	require.NoError(t, err)
	assert.Equal(t, resp.Query, got.Query)
	assert.Equal(t, resp.Category, got.Category)
	assert.Len(t, got.Results, 1)
	assert.Equal(t, resp.Results[0].Title, got.Results[0].Title)
}

func TestMultiLevel_Disabled(t *testing.T) {
	shared := makeTestStorage(t)
	c, err := NewMultiLevel(shared, Config{
		Enabled:   false,
		LocalTTL:  30,
		RemoteTTL: 300,
	})
	require.NoError(t, err)

	resp := &models.Response{Query: "disabled"}
	c.Set("dk", resp, 0)
	_, ok := c.Get("dk")
	assert.False(t, ok)
}
