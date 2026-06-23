package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/pkg/models"
)

// stubEngine is a minimal Engine implementation for testing the Loader.
type stubEngine struct {
	name       string
	categories []models.Category
	caps       Capabilities
	setupOk    bool
	initOk     bool
	setupCount int
	initCount  int
}

func (s *stubEngine) Name() string                   { return s.name }
func (s *stubEngine) Categories() []models.Category   { return s.categories }
func (s *stubEngine) Capabilities() Capabilities      { return s.caps }
func (s *stubEngine) About() EngineAbout              { return EngineAbout{} }
func (s *stubEngine) Setup(cfg EngineInitConfig) bool {
	s.setupCount++
	return s.setupOk
}
func (s *stubEngine) Init(ctx context.Context, cfg EngineInitConfig) bool {
	s.initCount++
	return s.initOk
}
func (s *stubEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return &models.Response{}, nil
}

func TestLoadEngines_Success(t *testing.T) {
	Reset()
	eng := &stubEngine{name: "testeng", categories: []models.Category{models.CategoryGeneral}, setupOk: true, initOk: true}
	Register("testeng", eng)

	cfgs := []EngineInitConfig{
		{Name: "testeng", Shortcut: "te"},
	}

	loader := NewLoader(nil) // no traits
	result, err := loader.Load(context.Background(), cfgs)
	require.NoError(t, err)
	require.NotNil(t, result)
	loader.Wait(2 * time.Second)
	assert.Len(t, result.Categories, 1)
	assert.Contains(t, result.Categories, "general")
	assert.Len(t, result.Shortcuts, 1)
	assert.Equal(t, "testeng", result.Shortcuts["te"])
}

func TestLoadEngines_NotFound(t *testing.T) {
	cfgs := []EngineInitConfig{{Name: "nonexistent"}}
	loader := NewLoader(nil)
	_, err := loader.Load(context.Background(), cfgs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoadEngines_SetupFails_MarksInactive(t *testing.T) {
	Reset()
	eng := &stubEngine{name: "badsetup", setupOk: false}
	Register("badsetup", eng)

	loader := NewLoader(nil)
	result, err := loader.Load(context.Background(), []EngineInitConfig{{Name: "badsetup"}})
	require.NoError(t, err)
	// After Load, the engine should NOT be in registry
	_, ok := Get("badsetup")
	assert.False(t, ok, "engine with failed setup should not be in registry")
	_ = result
}

func TestLoadEngines_NameValidation(t *testing.T) {
	loader := NewLoader(nil)

	err := loader.validateName("google-images")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lowercase alphanumeric")

	err = loader.validateName("Google")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lowercase")

	assert.NoError(t, loader.validateName("google"))
	assert.NoError(t, loader.validateName("google_images"))
	assert.NoError(t, loader.validateName("wikipedia"))
}

func TestLoadEngines_DuplicateName(t *testing.T) {
	cfgs := []EngineInitConfig{
		{Name: "dup", Shortcut: "d1"},
		{Name: "dup", Shortcut: "d2"},
	}

	loader := NewLoader(nil)
	_, err := loader.Load(context.Background(), cfgs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestLoadEngines_DuplicateShortcut(t *testing.T) {
	Reset()
	eng1 := &stubEngine{name: "eng1", setupOk: true, initOk: true}
	eng2 := &stubEngine{name: "eng2", setupOk: true, initOk: true}
	Register("eng1", eng1)
	Register("eng2", eng2)

	cfgs := []EngineInitConfig{
		{Name: "eng1", Shortcut: "same"},
		{Name: "eng2", Shortcut: "same"},
	}

	loader := NewLoader(nil)
	_, err := loader.Load(context.Background(), cfgs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate shortcut")
}

func TestLoadEngines_EmptyName(t *testing.T) {
	cfgs := []EngineInitConfig{{Name: ""}}

	loader := NewLoader(nil)
	_, err := loader.Load(context.Background(), cfgs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestLoadEngines_SkipInactive(t *testing.T) {
	Reset()
	eng := &stubEngine{name: "inactiveeng", setupOk: true, initOk: true}
	Register("inactiveeng", eng)

	cfgs := []EngineInitConfig{{Name: "inactiveeng", Inactive: true}}

	loader := NewLoader(nil)
	_, err := loader.Load(context.Background(), cfgs)
	require.NoError(t, err)
	_, ok := Get("inactiveeng")
	assert.False(t, ok, "inactive engine should be skipped")
}

func TestLoadEngines_Shortcut(t *testing.T) {
	Reset()
	eng := &stubEngine{name: "shortcuteng", setupOk: true, initOk: true}
	Register("shortcuteng", eng)

	cfgs := []EngineInitConfig{{Name: "shortcuteng", Shortcut: "se"}}

	loader := NewLoader(nil)
	result, err := loader.Load(context.Background(), cfgs)
	require.NoError(t, err)
	loader.Wait(2 * time.Second)
	assert.Equal(t, "shortcuteng", result.Shortcuts["se"])
}
