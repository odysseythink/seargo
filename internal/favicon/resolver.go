package favicon

import (
	"context"
	"fmt"
	"sync"
)

// ResolverFunc resolves a favicon for the given authority.
type ResolverFunc func(ctx context.Context, authority string) ([]byte, string, error)

var (
	mu        sync.RWMutex
	resolvers = map[string]ResolverFunc{}
)

// Register registers a favicon resolver by name.
func Register(name string, fn ResolverFunc) {
	mu.Lock()
	defer mu.Unlock()
	resolvers[name] = fn
}

// GetResolver returns a registered resolver by name.
func GetResolver(name string) (ResolverFunc, error) {
	mu.RLock()
	defer mu.RUnlock()
	fn, ok := resolvers[name]
	if !ok {
		return nil, fmt.Errorf("favicon resolver %q not found", name)
	}
	return fn, nil
}
