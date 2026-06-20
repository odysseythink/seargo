package plugin

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/seargo/seargo/pkg/models"
)

// PluginInfo describes a plugin for the preferences UI.
type PluginInfo struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	PreferenceSection string   `json:"preference_section"` // "general", "ui", "privacy", "query"
	Examples          []string `json:"examples,omitempty"`
	Keywords          []string `json:"keywords,omitempty"`
}

// Plugin is the interface that every SearGo plugin must implement.
type Plugin interface {
	ID() string
	Info() PluginInfo
	Init(ctx *AppContext) bool
	PreSearch(ctx *SearchContext) bool
	OnResult(ctx *SearchContext, result *models.Result) bool
	PostSearch(ctx *SearchContext) []models.Result
}

// PluginStorage manages the lifecycle of all registered plugins.
type PluginStorage struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewPluginStorage creates an empty plugin storage.
func NewPluginStorage() *PluginStorage {
	return &PluginStorage{
		plugins: make(map[string]Plugin),
	}
}

var (
	storageMu    sync.RWMutex
	globalPlugin *PluginStorage
)

// SetGlobalPlugin sets the global plugin storage.
func SetGlobalPlugin(ps *PluginStorage) {
	storageMu.Lock()
	defer storageMu.Unlock()
	globalPlugin = ps
}

// GlobalPlugin returns the global plugin storage.
func GlobalPlugin() *PluginStorage {
	storageMu.RLock()
	defer storageMu.RUnlock()
	return globalPlugin
}

// ResetForTest clears the global plugin storage (tests only).
func ResetForTest() {
	storageMu.Lock()
	defer storageMu.Unlock()
	globalPlugin = nil
}

// Register adds a plugin. Returns error on ID collision.
func (ps *PluginStorage) Register(p Plugin) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	id := p.ID()
	if _, exists := ps.plugins[id]; exists {
		return fmt.Errorf("plugin: id %q already registered", id)
	}
	ps.plugins[id] = p
	return nil
}

// Get retrieves a plugin by ID.
func (ps *PluginStorage) Get(id string) (Plugin, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.plugins[id]
	return p, ok
}

// All returns all registered plugins.
func (ps *PluginStorage) All() []Plugin {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	result := make([]Plugin, 0, len(ps.plugins))
	for _, p := range ps.plugins {
		result = append(result, p)
	}
	return result
}

// PreSearch runs pre_search hooks for plugins in the user's enabled list.
// Returns false if any plugin aborts the search.
func (ps *PluginStorage) PreSearch(ctx *SearchContext) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, id := range ctx.UserPlugins {
		p, ok := ps.plugins[id]
		if !ok {
			continue
		}
		if !callPreSearch(p, ctx) {
			return false
		}
	}
	return true
}

// OnResult runs on_result hooks for each enabled plugin on a single result.
// Returns false if any plugin removes the result.
func (ps *PluginStorage) OnResult(ctx *SearchContext, r *models.Result) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, id := range ctx.UserPlugins {
		p, ok := ps.plugins[id]
		if !ok {
			continue
		}
		if !callOnResult(p, ctx, r) {
			return false
		}
	}
	return true
}

// PostSearch runs post_search hooks and collects results.
// Plugin keywords are matched against the first word of the query.
func (ps *PluginStorage) PostSearch(ctx *SearchContext) []models.Result {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var all []models.Result
	for _, id := range ctx.UserPlugins {
		p, ok := ps.plugins[id]
		if !ok {
			continue
		}
		info := p.Info()
		if len(info.Keywords) > 0 {
			firstWord := firstWordOf(ctx.Query)
			if !containsStr(info.Keywords, firstWord) {
				continue
			}
		}
		results := callPostSearch(p, ctx)
		all = append(all, results...)
	}
	return all
}

// --- panic-safe call wrappers ---

func callPreSearch(p Plugin, ctx *SearchContext) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			logPanic("plugin", p.ID(), "pre_search", r)
			ok = true
		}
	}()
	return p.PreSearch(ctx)
}

func callOnResult(p Plugin, ctx *SearchContext, r *models.Result) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			logPanic("plugin", p.ID(), "on_result", r)
			ok = true
		}
	}()
	return p.OnResult(ctx, r)
}

func callPostSearch(p Plugin, ctx *SearchContext) (results []models.Result) {
	defer func() {
		if r := recover(); r != nil {
			logPanic("plugin", p.ID(), "post_search", r)
			results = nil
		}
	}()
	return p.PostSearch(ctx)
}

// --- helpers ---

func firstWordOf(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, ' '); idx > 0 {
		return strings.ToLower(s[:idx])
	}
	return strings.ToLower(s)
}

func containsStr(list []string, s string) bool {
	s = strings.ToLower(s)
	for _, item := range list {
		if strings.ToLower(item) == s {
			return true
		}
	}
	return false
}

func logPanic(kind, id, hook string, recovered any) {
	fmt.Fprintf(os.Stderr, "[seargo] panic in %s %s.%s: %v\n", kind, id, hook, recovered)
}
