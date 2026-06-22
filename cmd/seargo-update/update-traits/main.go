package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/seargo/seargo/cmd/seargo-update/internal"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
)

const defaultConfigPath = "configs/settings.yml"

// traitFetcher fetches language/region traits for a single engine.
type traitFetcher func(ctx context.Context, h *fetch.Helper, cfg config.EngineConfig) (engine.EngineTraits, error)

var traitFetchers = map[string]traitFetcher{}

func registerTraitFetcher(name string, fn traitFetcher) {
	traitFetchers[name] = fn
}

func main() {
	var (
		out        = flag.String("out", "data/engine_traits.json", "output JSON path")
		configPath = flag.String("config", defaultConfigPath, "settings YAML path")
	)
	flag.Parse()

	if err := Run(*out, nil, *configPath, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "update-traits: %v\n", err)
		os.Exit(1)
	}
}

// Run updates engine traits. If engines is empty, all engines from the config
// are considered. Engines without a registered fetcher keep their existing
// traits (when the output file exists) to avoid data loss.
func Run(outPath string, client fetch.Client, configPath string, engines []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	traits, err := loadExistingTraits(outPath)
	if err != nil {
		return err
	}
	if traits == nil {
		traits = make(engine.EngineTraitsMap)
	}

	h := fetch.New(client)
	ctx := context.Background()

	filter := make(map[string]bool)
	for _, name := range engines {
		filter[name] = true
	}
	hasFilter := len(filter) > 0

	for _, ec := range cfg.Engines {
		if ec.Inactive {
			continue
		}
		name := ec.Name
		if name == "" {
			name = ec.Engine
		}
		if hasFilter && !filter[name] {
			continue
		}

		fetcher, ok := traitFetchers[name]
		if !ok {
			// Preserve existing traits when no fetcher is registered.
			continue
		}
		ec.Inactive = false
		t, err := fetcher(ctx, h, ec)
		if err != nil {
			return fmt.Errorf("fetch traits for %s: %w", name, err)
		}
		traits[name] = t
	}

	return writeJSON(outPath, traits)
}

func loadExistingTraits(path string) (engine.EngineTraitsMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(engine.EngineTraitsMap), nil
		}
		return nil, fmt.Errorf("read existing traits: %w", err)
	}
	var tm engine.EngineTraitsMap
	if err := json.Unmarshal(data, &tm); err != nil {
		return nil, fmt.Errorf("parse existing traits: %w", err)
	}
	return tm, nil
}

func writeJSON(outPath string, traits engine.EngineTraitsMap) error {
	// Re-marshal through a sorted map to keep deterministic output.
	enc, err := json.MarshalIndent(sortedTraits(traits), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	enc = append(enc, '\n')

	tmp := outPath + ".tmp"
	if err := os.WriteFile(tmp, enc, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	return os.Rename(tmp, outPath)
}

func sortedTraits(traits engine.EngineTraitsMap) map[string]engine.EngineTraits {
	keys := make([]string, 0, len(traits))
	for k := range traits {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]engine.EngineTraits, len(keys))
	for _, k := range keys {
		ordered[k] = traits[k]
	}
	return ordered
}
