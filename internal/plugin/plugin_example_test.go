package plugin

import (
	"net/rpc"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildExamplePlugin(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	name := "echo"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binPath := filepath.Join(dir, name)

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/plugin-example")
	cmd.Dir = "../.." // repo root relative to internal/plugin
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "build example plugin: %s", out)
	return binPath
}

func TestExamplePlugin_PostSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping plugin binary build in short mode")
	}

	binPath := buildExamplePlugin(t)

	clientConfig := plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          map[string]plugin.Plugin{"external_plugin": &ExternalPluginPlugin{Impl: nil}},
		Cmd:              exec.Command(binPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
		Logger:           hclog.NewNullLogger(),
	}

	client := plugin.NewClient(&clientConfig)
	defer client.Kill()

	rpcClient, err := client.Client()
	require.NoError(t, err)

	raw, err := rpcClient.Dispense("external_plugin")
	require.NoError(t, err)

	netRPC, ok := raw.(*rpc.Client)
	require.True(t, ok)
	defer netRPC.Close()

	args := PostSearchArgs{Ctx: SearchContext{Query: "hello world"}}
	var reply PostSearchReply
	require.NoError(t, netRPC.Call("Plugin.PostSearch", &args, &reply))

	require.Len(t, reply.Results, 1)
	assert.Equal(t, "answer", reply.Results[0].Kind)
	assert.Equal(t, "Hello world", reply.Results[0].Content)
}
