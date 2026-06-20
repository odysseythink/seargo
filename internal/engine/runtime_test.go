package engine

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/pkg/models"
)

type lifecycleEngine struct {
	name       string
	setupOk    bool
	initDelay  time.Duration
	initOk     atomic.Bool
	initCalled atomic.Bool
}

func (e *lifecycleEngine) Name() string                  { return e.name }
func (e *lifecycleEngine) Categories() []models.Category  { return nil }
func (e *lifecycleEngine) Capabilities() Capabilities     { return Capabilities{} }
func (e *lifecycleEngine) About() EngineAbout             { return EngineAbout{} }

func (e *lifecycleEngine) Setup(cfg EngineInitConfig) bool {
	return e.setupOk
}

func (e *lifecycleEngine) Init(ctx context.Context, cfg EngineInitConfig) bool {
	e.initCalled.Store(true)
	if e.initDelay > 0 {
		select {
		case <-time.After(e.initDelay):
		case <-ctx.Done():
			return false
		}
	}
	e.initOk.Store(true)
	return true
}

func (e *lifecycleEngine) IsInitOk() bool {
	return e.initOk.Load()
}

func (e *lifecycleEngine) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	return &models.Response{}, nil
}

func TestLifecycle_SetupFailSkipsInit(t *testing.T) {
	Reset()
	eng := &lifecycleEngine{name: "failsetup", setupOk: false}
	Register("failsetup", eng)

	loader := NewLoader(nil)
	cfgs := []EngineInitConfig{{Name: "failsetup"}}
	result, err := loader.Load(context.Background(), cfgs)
	require.NoError(t, err)
	assert.NotNil(t, result)

	_, ok := Get("failsetup")
	assert.False(t, ok)
	assert.False(t, eng.initCalled.Load())
}

func TestLifecycle_AsyncInitCompletes(t *testing.T) {
	Reset()
	eng := &lifecycleEngine{name: "asyncok", setupOk: true, initDelay: 50 * time.Millisecond}
	Register("asyncok", eng)

	loader := NewLoader(nil)
	cfgs := []EngineInitConfig{{Name: "asyncok"}}
	_, err := loader.Load(context.Background(), cfgs)
	require.NoError(t, err)

	loader.Wait(5 * time.Second)

	assert.True(t, eng.IsInitOk())
}

func TestLifecycle_InitFailsMarksInactive(t *testing.T) {
	Reset()
	eng := &lifecycleEngine{name: "initfail", setupOk: true, initDelay: 200 * time.Millisecond}
	Register("initfail", eng)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	loader := NewLoader(nil)
	cfgs := []EngineInitConfig{{Name: "initfail"}}
	_, err := loader.Load(ctx, cfgs)
	require.NoError(t, err)

	loader.Wait(5 * time.Second)

	_, ok := Get("initfail")
	assert.False(t, ok, "engine with failed init should not be in registry")
}

func TestLifecycle_HotReload_NewRegistry(t *testing.T) {
	Reset()
	eng1 := &lifecycleEngine{name: "eng1", setupOk: true, initDelay: 10 * time.Millisecond}
	Register("eng1", eng1)
	eng2 := &lifecycleEngine{name: "eng2", setupOk: true, initDelay: 10 * time.Millisecond}
	Register("eng2", eng2)

	// Load eng1
	loader := NewLoader(nil)
	_, err := loader.Load(context.Background(), []EngineInitConfig{{Name: "eng1"}})
	require.NoError(t, err)
	loader.Wait(5 * time.Second)

	_, ok := Get("eng1")
	assert.True(t, ok)

	// Reload with eng2 only (re-register because Reset clears the registry)
	Register("eng2", eng2)
	_, err = loader.Load(context.Background(), []EngineInitConfig{{Name: "eng2"}})
	require.NoError(t, err)
	loader.Wait(5 * time.Second)

	_, ok = Get("eng1")
	assert.False(t, ok)
	_, ok = Get("eng2")
	assert.True(t, ok)
}
