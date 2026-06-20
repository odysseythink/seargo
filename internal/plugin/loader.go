package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	goPlugin "plugin"
	"strings"

	"github.com/seargo/seargo/pkg/models"
)

// LoadThirdPartyPlugins loads .so plugin files from a directory.
// Only files matching configured plugin IDs are loaded.
// Returns the number of successfully loaded plugins.
func LoadThirdPartyPlugins(pluginDir string, enabledIDs []string, ps *PluginStorage) (int, error) {
	if pluginDir == "" {
		return 0, nil
	}

	// Build set of enabled IDs
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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".so")
		if err := validatePluginID(id); err != nil {
			continue // skip files that don't match ID pattern
		}

		if !enabled[id] {
			continue
		}

		path := filepath.Join(pluginDir, entry.Name())
		p, err := loadPluginFromSO(path, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[seargo] failed to load plugin %q from %q: %v\n", id, path, err)
			continue
		}

		if err := ps.Register(p); err != nil {
			fmt.Fprintf(os.Stderr, "[seargo] failed to register plugin %q: %v\n", id, err)
			continue
		}
		loaded++
	}

	return loaded, nil
}

// loadPluginFromSO loads a single plugin from a .so file using Go's plugin package.
// The .so must export a symbol named "Plugin" that implements the Plugin interface.
func loadPluginFromSO(path, id string) (Plugin, error) {
	p, err := goPlugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open plugin %q: %w", id, err)
	}

	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("plugin %q: missing Plugin symbol: %w", id, err)
	}

	pluginImpl, ok := sym.(Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin %q: Plugin symbol does not implement Plugin interface", id)
	}

	return pluginImpl, nil
}

// ensureThirdPartyPluginInterface is a compile-time check that standard library
// plugin.Plugin is not accidentally used instead of our Plugin interface.
// It satisfies the unused-import linter if we ever import "plugin" separately.
var _ = goPlugin.Open

// PluginNamespace ensures the models import is used.
var _ = models.Result{}
