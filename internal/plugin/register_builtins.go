package plugin

import (
	"fmt"
	"regexp"
	"sync"
)

// builtinFactory creates a Plugin instance.
type builtinFactory func() Plugin

var (
	builtinMu    sync.RWMutex
	builtinRegs  = make(map[string]builtinFactory)
)

// RegisterBuiltin registers a built-in plugin factory. Called from init().
// Panics if the id is invalid or already registered.
func RegisterBuiltin(id string, factory builtinFactory) {
	if err := validatePluginID(id); err != nil {
		panic(fmt.Errorf("plugin: invalid built-in id %q: %w", id, err))
	}
	builtinMu.Lock()
	defer builtinMu.Unlock()
	if _, exists := builtinRegs[id]; exists {
		panic(fmt.Errorf("plugin: built-in %q already registered", id))
	}
	builtinRegs[id] = factory
}

// BuiltinRegistrations returns a copy of the built-in registration map.
func BuiltinRegistrations() map[string]builtinFactory {
	builtinMu.RLock()
	defer builtinMu.RUnlock()
	result := make(map[string]builtinFactory, len(builtinRegs))
	for k, v := range builtinRegs {
		result[k] = v
	}
	return result
}

// RegisterBuiltinsFromList instantiates and registers all built-in plugins
// from the given registrations into the provided storage.
func RegisterBuiltinsFromList(ps *PluginStorage, registrations map[string]builtinFactory) {
	for id, factory := range registrations {
		p := factory()
		if err := ps.Register(p); err != nil {
			logPanic("plugin", id, "register", err)
		}
	}
}

// idPattern matches valid plugin IDs: lowercase alphanumeric + underscore only, must start with a letter.
var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// validatePluginID checks that a plugin ID is valid.
func validatePluginID(id string) error {
	if id == "" {
		return fmt.Errorf("plugin ID must not be empty")
	}
	if !idPattern.MatchString(id) {
		return fmt.Errorf("plugin ID %q must match %s", id, idPattern.String())
	}
	return nil
}
