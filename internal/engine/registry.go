package engine

import "sync"

var (
	registry = make(map[string]Engine)
	mu       sync.RWMutex
)

func Register(name string, e Engine) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = e
}

func Get(name string) (Engine, bool) {
	mu.RLock()
	defer mu.RUnlock()
	e, ok := registry[name]
	return e, ok
}

func All() map[string]Engine {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]Engine, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}

func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// SetAll replaces the entire registry atomically. Used by Loader during
// initialization and hot reload.
func SetAll(m map[string]Engine) {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]Engine, len(m))
	for k, v := range m {
		registry[k] = v
	}
}

// Reset clears the registry. Used in tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]Engine)
}
