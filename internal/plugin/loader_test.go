package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
