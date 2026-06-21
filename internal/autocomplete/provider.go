package autocomplete

import (
	"context"
	"sync"
)

// Provider fetches autocomplete suggestions from an external backend.
type Provider interface {
	Fetch(ctx context.Context, query string, locale string) ([]string, error)
}

var (
	providers   = make(map[string]Provider)
	providersMu sync.RWMutex
)

func Register(name string, p Provider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	if _, ok := providers[name]; ok {
		panic("autocomplete: duplicate provider registration: " + name)
	}
	providers[name] = p
}

func Get(name string) (Provider, bool) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	p, ok := providers[name]
	return p, ok
}

func All() map[string]Provider {
	providersMu.RLock()
	defer providersMu.RUnlock()
	result := make(map[string]Provider, len(providers))
	for k, v := range providers {
		result[k] = v
	}
	return result
}

func Names() []string {
	providersMu.RLock()
	defer providersMu.RUnlock()
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names
}

func Reset() {
	providersMu.Lock()
	defer providersMu.Unlock()
	providers = make(map[string]Provider)
}
