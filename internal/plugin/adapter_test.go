package plugin

import (
	"net"
	"net/rpc"
	"testing"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExternalPlugin is a test implementation of ExternalPlugin.
type mockExternalPlugin struct {
	id            string
	info          PluginInfo
	initOk        bool
	preSearchOk   bool
	onResultKeep  bool
	postSearchOut []models.Result
}

func (m *mockExternalPlugin) ID() string                                 { return m.id }
func (m *mockExternalPlugin) Info() PluginInfo                           { return m.info }
func (m *mockExternalPlugin) Init(map[string]any) bool                   { return m.initOk }
func (m *mockExternalPlugin) PreSearch(SearchContext) bool               { return m.preSearchOk }
func (m *mockExternalPlugin) OnResult(SearchContext, models.Result) bool { return m.onResultKeep }
func (m *mockExternalPlugin) PostSearch(SearchContext) []models.Result   { return m.postSearchOut }

// startMockRPCServer starts an in-memory net/rpc server backed by mockExternalPlugin.
func startMockRPCServer(t *testing.T, impl ExternalPlugin) *rpc.Client {
	t.Helper()
	srv := rpc.NewServer()
	require.NoError(t, srv.RegisterName("Plugin", &ExternalPluginRPC{Impl: impl}))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() })

	go srv.Accept(listener)

	client, err := rpc.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	return client
}

func TestExternalPluginAdapter_RPCSuccess(t *testing.T) {
	impl := &mockExternalPlugin{
		id:           "mock",
		info:         PluginInfo{ID: "mock", Name: "Mock"},
		initOk:       true,
		preSearchOk:  false,
		onResultKeep: false,
		postSearchOut: []models.Result{{Title: "extra", Content: "from plugin"}},
	}
	rpcClient := startMockRPCServer(t, impl)
	adapter := newExternalPluginAdapter("mock", impl.Info(), nil, rpcClient)

	ctx := &SearchContext{Query: "hello"}
	assert.True(t, adapter.Init(&AppContext{}))
	assert.False(t, adapter.PreSearch(ctx))
	assert.False(t, adapter.OnResult(ctx, &models.Result{Title: "x"}))
	results := adapter.PostSearch(ctx)
	require.Len(t, results, 1)
	assert.Equal(t, "from plugin", results[0].Content)
}

func TestExternalPluginAdapter_RPCFailure_ReturnsSafeDefault(t *testing.T) {
	// Use a real mock server, then break the connection.
	impl := &mockExternalPlugin{id: "mock", info: PluginInfo{ID: "mock"}}
	rpcClient := startMockRPCServer(t, impl)

	// Close the client to simulate a broken connection.
	rpcClient.Close()

	adapter := newExternalPluginAdapter("broken", PluginInfo{ID: "broken"}, nil, rpcClient)
	ctx := &SearchContext{Query: "hello"}

	assert.True(t, adapter.PreSearch(ctx), "PreSearch must return true on RPC failure")
	assert.True(t, adapter.OnResult(ctx, &models.Result{Title: "x"}), "OnResult must return true on RPC failure")
	assert.Nil(t, adapter.PostSearch(ctx), "PostSearch must return nil on RPC failure")

	// After a failure the adapter should be marked unavailable and skip RPC.
	assert.True(t, adapter.PreSearch(ctx))
}

func TestExtractPluginExtra(t *testing.T) {
	cfg := &config.Config{
		Plugins: map[string]config.PluginConfig{
			"echo": {
				Active: true,
				Extra:  map[string]any{"prefix": "> "},
			},
		},
	}
	appCtx := &AppContext{Config: cfg}

	assert.Equal(t, map[string]any{"prefix": "> "}, extractPluginExtra(appCtx, "echo"))
	assert.Nil(t, extractPluginExtra(appCtx, "unknown"))
	assert.Nil(t, extractPluginExtra(&AppContext{Config: "not-a-config"}, "echo"))
	assert.Nil(t, extractPluginExtra(nil, "echo"))
}
