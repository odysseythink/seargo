package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

// mockEngine is a minimal engine implementation for testing.
type mockEngine struct {
	name string
}

func (m *mockEngine) Name() string                       { return m.name }
func (m *mockEngine) Categories() []models.Category      { return nil }
func (m *mockEngine) Capabilities() Capabilities         { return Capabilities{} }
func (m *mockEngine) Init(client *httpx.Client, cfg EngineInitConfig) error { return nil }
func (m *mockEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return nil, nil
}

func TestRegistry(t *testing.T) {
	// Clear registry before test
	registry = make(map[string]Engine)

	e1 := &mockEngine{name: "google"}
	e2 := &mockEngine{name: "bing"}

	Register("google", e1)
	Register("bing", e2)

	got, ok := Get("google")
	assert.True(t, ok)
	assert.Equal(t, "google", got.Name())

	_, ok = Get("unknown")
	assert.False(t, ok)

	all := All()
	assert.Len(t, all, 2)

	names := Names()
	assert.Len(t, names, 2)
}

func TestCapabilitiesAndInfo(t *testing.T) {
	caps := Capabilities{
		SupportsSafeSearch: true,
		SupportsLanguage:   true,
		SupportsTimeRange:  false,
		SupportsPagination: true,
		RequiresAPIKey:     false,
		Shortcut:           "g",
	}
	info := Info{
		Name:         "google",
		Categories:   []string{"general", "images"},
		Shortcut:     "g",
		Capabilities: caps,
		Enabled:      true,
	}
	assert.Equal(t, "google", info.Name)
	assert.Equal(t, "g", info.Shortcut)
	assert.Equal(t, "g", info.Capabilities.Shortcut)
	assert.True(t, info.Capabilities.SupportsPagination)
	assert.False(t, info.Capabilities.SupportsTimeRange)
}
