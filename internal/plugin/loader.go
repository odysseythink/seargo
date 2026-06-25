package plugin

import (
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// startExternalPluginFn is overridable in tests.
var startExternalPluginFn = startExternalPlugin

// LoadThirdPartyPlugins discovers and launches external plugin executables
// from pluginDir. Only plugins whose IDs are in enabledIDs are loaded.
// Returns the number of successfully registered plugins.
func LoadThirdPartyPlugins(pluginDir string, enabledIDs []string, ps *PluginStorage) (int, error) {
	if pluginDir == "" {
		return 0, nil
	}

	enabled := make(map[string]bool, len(enabledIDs))
	for _, id := range enabledIDs {
		enabled[id] = true
	}

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return 0, fmt.Errorf("read plugin dir %q: %w", pluginDir, err)
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		id, ok := parseExecutableName(entry.Name(), runtime.GOOS)
		if !ok {
			continue
		}

		if err := validatePluginID(id); err != nil {
			fmt.Fprintf(os.Stderr, "[seargo] invalid plugin id %q, skipping\n", id)
			continue
		}

		if !enabled[id] {
			continue
		}

		path := filepath.Join(pluginDir, entry.Name())
		p, err := startExternalPluginFn(path, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[seargo] failed to start plugin %q from %q: %v\n", id, path, err)
			continue
		}

		if err := ps.Register(p); err != nil {
			if closer, ok := p.(interface{ Close() error }); ok {
				closer.Close()
			}
			fmt.Fprintf(os.Stderr, "[seargo] failed to register plugin %q: %v\n", id, err)
			continue
		}
		loaded++
	}

	return loaded, nil
}

// parseExecutableName extracts a plugin ID from a directory entry name.
// On Windows the name must end with ".exe"; on Unix it must contain no dot.
func parseExecutableName(name, goos string) (string, bool) {
	var id string
	if goos == "windows" {
		if !strings.HasSuffix(name, ".exe") {
			return "", false
		}
		id = strings.TrimSuffix(name, ".exe")
	} else {
		if strings.Contains(name, ".") {
			return "", false
		}
		id = name
	}
	if id == "" {
		return "", false
	}
	return id, true
}

// startExternalPlugin launches a single external plugin executable and returns
// an adapter that satisfies the internal Plugin interface.
func startExternalPlugin(path, id string) (Plugin, error) {
	clientConfig := plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          map[string]plugin.Plugin{"external_plugin": &ExternalPluginPlugin{Impl: nil}},
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
		Logger:           hclog.NewNullLogger(),
	}

	client := plugin.NewClient(&clientConfig)
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("plugin %q handshake failed: %w", id, err)
	}

	raw, err := rpcClient.Dispense("external_plugin")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("plugin %q dispense failed: %w", id, err)
	}

	netRPC, ok := raw.(*rpc.Client)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin %q returned unexpected client type %T", id, raw)
	}

	// Retrieve plugin metadata via RPC. The plugin's Info() does not depend on Init.
	var infoReply struct {
		Info PluginInfo
	}
	if err := netRPC.Call("Plugin.Info", &InfoArgs{}, &infoReply); err != nil {
		client.Kill()
		return nil, fmt.Errorf("plugin %q Info RPC failed: %w", id, err)
	}

	adapter := newExternalPluginAdapter(id, infoReply.Info, client, netRPC)
	return adapter, nil
}
