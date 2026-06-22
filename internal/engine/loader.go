package engine

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// engineNamePattern validates engine names: lowercase alphanumeric, no underscore.
var engineNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// Loader validates engine configs, resolves traits, calls Setup/Init,
// and builds a validated Registry with category and shortcut maps.
type Loader struct {
	traits     EngineTraitsMap
	wg         sync.WaitGroup
	initCtx    context.Context
	initCancel context.CancelFunc
}

// NewLoader creates a new Loader.
func NewLoader(traits EngineTraitsMap) *Loader {
	if traits == nil {
		traits = make(EngineTraitsMap)
	}
	return &Loader{traits: traits}
}

// LoadResult holds the output of a successful Load.
type LoadResult struct {
	Categories map[string][]string // category → list of engine names
	Shortcuts  map[string]string   // shortcut → engine name
}

// Load validates configs, runs synchronous Setup, and starts async Init
// goroutines for each engine. The registry is cleared before the async
// phase so that only successfully initialized engines from this call
// are present after Wait completes (supports hot reload).
func (l *Loader) Load(ctx context.Context, configs []EngineInitConfig) (*LoadResult, error) {
	l.initCtx, l.initCancel = context.WithCancel(ctx)

	if err := l.validateConfigs(configs); err != nil {
		return nil, err
	}

	type pendingEngine struct {
		eng Engine
		cfg EngineInitConfig
	}

	categories := make(map[string][]string)
	shortcuts := make(map[string]string)

	// Step 1: Collect engines from the registry before clearing it.
	var pending []pendingEngine

	for _, cfg := range configs {
		if cfg.Inactive {
			continue
		}

		eng, ok := Get(cfg.Name)
		if !ok {
			return nil, fmt.Errorf("engine %q not found in registry", cfg.Name)
		}

		traits, _ := l.traits.Lookup(cfg.Name)
		cfg.EngineTraits = traits

		if !eng.Setup(cfg) {
			continue
		}

		pending = append(pending, pendingEngine{eng: eng, cfg: cfg})

		for _, cat := range eng.Categories() {
			catStr := string(cat)
			categories[catStr] = append(categories[catStr], cfg.Name)
		}
		if len(eng.Categories()) == 0 {
			categories["other"] = append(categories["other"], cfg.Name)
		}
		if cfg.Shortcut != "" {
			shortcuts[cfg.Shortcut] = cfg.Name
		}
	}

	// Step 2: Clear the registry so only engines that succeed in
	// async Init remain (supports hot reload).
	Reset()

	// Step 3: Start async Init goroutines
	for _, p := range pending {
		l.wg.Add(1)
		go func(eng Engine, cfg EngineInitConfig) {
			defer l.wg.Done()
			if eng.Init(l.initCtx, cfg) {
				Register(cfg.Name, eng)
			}
		}(p.eng, p.cfg)
	}

	return &LoadResult{
		Categories: categories,
		Shortcuts:  shortcuts,
	}, nil
}

// Wait blocks until all async Init goroutines complete or timeout.
func (l *Loader) Wait(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		l.initCancel()
		return false
	}
}

// Shutdown cancels all pending Init goroutines.
func (l *Loader) Shutdown() {
	if l.initCancel != nil {
		l.initCancel()
	}
	l.wg.Wait()
}

// validateConfigs checks for name validity, duplicates, and shortcut collisions.
func (l *Loader) validateConfigs(configs []EngineInitConfig) error {
	seen := make(map[string]bool)
	shortcuts := make(map[string]string)

	for i, cfg := range configs {
		if cfg.Name == "" {
			return fmt.Errorf("engine[%d]: name is empty", i)
		}
		if err := l.validateName(cfg.Name); err != nil {
			return fmt.Errorf("engine[%d] %q: %w", i, cfg.Name, err)
		}
		lower := strings.ToLower(cfg.Name)
		if seen[lower] {
			return fmt.Errorf("engine[%d] %q: duplicate engine name", i, cfg.Name)
		}
		seen[lower] = true
		if cfg.Shortcut != "" {
			if existing, ok := shortcuts[cfg.Shortcut]; ok {
				return fmt.Errorf("engine[%d] %q: duplicate shortcut %q (already used by %s)",
					i, cfg.Name, cfg.Shortcut, existing)
			}
			shortcuts[cfg.Shortcut] = cfg.Name
		}
	}
	return nil
}

// validateName checks engine name rules: lowercase, no underscore.
func (l *Loader) validateName(name string) error {
	if !engineNamePattern.MatchString(name) {
		return fmt.Errorf("engine name must be lowercase alphanumeric without underscore, got %q", name)
	}
	return nil
}
