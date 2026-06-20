package deps

import "sync"

// AhmiaBlacklist is a concurrency-safe set of onion service hash blacklist entries.
type AhmiaBlacklist struct {
	mu     sync.RWMutex
	hashes map[string]bool
}

// NewAhmiaBlacklist creates a new empty AhmiaBlacklist.
func NewAhmiaBlacklist() *AhmiaBlacklist {
	return &AhmiaBlacklist{
		hashes: make(map[string]bool),
	}
}

// Add inserts a hash into the blacklist. It is safe for concurrent use.
func (b *AhmiaBlacklist) Add(hash string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.hashes[hash] = true
}

// Contains reports whether the given hash is in the blacklist.
func (b *AhmiaBlacklist) Contains(hash string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.hashes[hash]
}

// LoadFromHashes loads multiple hashes into the blacklist at once.
func (b *AhmiaBlacklist) LoadFromHashes(hashes []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, h := range hashes {
		b.hashes[h] = true
	}
}
