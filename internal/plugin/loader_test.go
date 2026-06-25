package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePluginID(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{"calculator", false},
		{"hash_plugin", false},
		{"myPlugin123", true},   // uppercase not allowed
		{"", true},
		{"123abc", true},        // must start with letter
		{"my-plugin", true},     // no hyphens
		{"my plugin", true},     // no spaces
		{"../evil", true},       // path traversal
		{"plugin.so", true},     // no dots
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			err := validatePluginID(tt.id)
			if tt.wantErr {
				assert.Error(t, err, "expected error for id=%q", tt.id)
			} else {
				assert.NoError(t, err, "expected no error for id=%q", tt.id)
			}
		})
	}
}

func TestRegisterBuiltin_And_BuiltinRegistrations(t *testing.T) {
	// Can't call RegisterBuiltin twice with same id - would panic.
	// Instead, test the registrations map directly.
	builtinMu.Lock()
	builtinRegs["test_builtin"] = func() Plugin {
		return &mockPlugin{id: "test_builtin"}
	}
	builtinMu.Unlock()

	regs := BuiltinRegistrations()
	assert.Contains(t, regs, "test_builtin")

	// Clean up
	builtinMu.Lock()
	delete(builtinRegs, "test_builtin")
	builtinMu.Unlock()
}

func TestRegisterBuiltinsFromList(t *testing.T) {
	ResetForTest()
	ps := NewPluginStorage()

	regs := map[string]builtinFactory{
		"mock_a": func() Plugin { return &mockPlugin{id: "mock_a"} },
		"mock_b": func() Plugin { return &mockPlugin{id: "mock_b"} },
	}

	RegisterBuiltinsFromList(ps, regs)
	assert.Len(t, ps.All(), 2)
}


func TestParseExecutableName(t *testing.T) {
	tests := []struct {
		name   string
		goos   string
		wantID string
		wantOK bool
	}{
		// must survive
		{"echo", "linux", "echo", true},
		{"echo", "darwin", "echo", true},
		{"echo.exe", "windows", "echo", true},
		{"my_plugin", "linux", "my_plugin", true},
		// must reject
		{"echo.so", "linux", "", false},
		{"echo.exe", "linux", "", false},
		{"my-plugin", "linux", "my-plugin", true}, // passes name parsing, rejected by validatePluginID later
		{"123abc", "linux", "123abc", true},       // passes name parsing, rejected by validatePluginID later
		{"", "linux", "", false},
		{"echo", "windows", "", false},
		{"echo.txt", "windows", "", false},
		{"bad-id.exe", "windows", "bad-id", true}, // passes name parsing, rejected by validatePluginID later
	}
	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.goos, func(t *testing.T) {
			id, ok := parseExecutableName(tt.name, tt.goos)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestLoadThirdPartyPlugins_SkipsDisabled(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "echo"), []byte(""), 0o755))

	ps := NewPluginStorage()
	loaded, err := LoadThirdPartyPlugins(dir, []string{"other"}, ps)
	require.NoError(t, err)
	assert.Equal(t, 0, loaded)
	assert.Empty(t, ps.All())
}

func TestLoadThirdPartyPlugins_StartsEnabled(t *testing.T) {
	origStartFn := startExternalPluginFn
	startExternalPluginFn = func(path, id string) (Plugin, error) {
		return &mockPlugin{id: id}, nil
	}
	defer func() { startExternalPluginFn = origStartFn }()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "echo"), []byte(""), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "invalid-id"), []byte(""), 0o755))

	ps := NewPluginStorage()
	loaded, err := LoadThirdPartyPlugins(dir, []string{"echo"}, ps)
	require.NoError(t, err)
	assert.Equal(t, 1, loaded)
	assert.Len(t, ps.All(), 1)
	assert.Equal(t, "echo", ps.All()[0].ID())
}

func TestLoadThirdPartyPlugins_PluginDirEmpty(t *testing.T) {
	ps := NewPluginStorage()
	loaded, err := LoadThirdPartyPlugins("", []string{"echo"}, ps)
	require.NoError(t, err)
	assert.Equal(t, 0, loaded)
}
