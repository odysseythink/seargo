package porting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SmokeTestConfig defines the configuration for running smoke tests.
type SmokeTestConfig struct {
	SearxngPath string // path to SearXNG source
	OutputDir   string // output directory for generated Go files
	BaseType    string // optional: filter by base type (xpath, json_engine, etc.)
	Limit       int    // max engines to generate (0 = all)
}

// SmokeResult holds the outcome of generating a single engine.
type SmokeResult struct {
	Name   string
	Path   string
	Error  string
	Output string
}

// RunSmoke read all Python engine files from the SearXNG source directory,
// generates Go skeletons for each, and writes them to the output directory.
func RunSmoke(cfg SmokeTestConfig) []SmokeResult {
	engineDir := filepath.Join(cfg.SearxngPath, "searx", "engines")

	entries, err := os.ReadDir(engineDir)
	if err != nil {
		return []SmokeResult{{
			Name:  "init",
			Error: fmt.Sprintf("read SearXNG engine dir: %v", err),
		}}
	}

	var results []SmokeResult
	count := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".py") {
			continue
		}

		if cfg.Limit > 0 && count >= cfg.Limit {
			break
		}

		pyPath := filepath.Join(engineDir, entry.Name())
		engineName := strings.TrimSuffix(entry.Name(), ".py")

		// Skip __init__.py and similar
		if strings.HasPrefix(engineName, "__") {
			continue
		}

		pySource, err := os.ReadFile(pyPath)
		if err != nil {
			results = append(results, SmokeResult{
				Name:  engineName,
				Error: fmt.Sprintf("read file: %v", err),
			})
			continue
		}

		result, err := GenerateSkeleton(engineName, string(pySource))
		if err != nil {
			results = append(results, SmokeResult{
				Name:  engineName,
				Error: err.Error(),
			})
			continue
		}

		if cfg.BaseType != "" && result.BaseType != cfg.BaseType {
			continue
		}

		// Write Go file
		goPath := filepath.Join(cfg.OutputDir, engineName+".go")
		if err := os.WriteFile(goPath, []byte(result.GoCode), 0644); err != nil {
			results = append(results, SmokeResult{
				Name:  engineName,
				Error: fmt.Sprintf("write Go file: %v", err),
			})
			continue
		}

		// Write fixture file
		fixtureDir := filepath.Join(cfg.OutputDir, "..", "testdata", "fixtures", "engines")
		os.MkdirAll(fixtureDir, 0755)
		fixturePath := filepath.Join(fixtureDir, engineName+".yaml")
		os.WriteFile(fixturePath, []byte(result.FixtureYAML), 0644)

		results = append(results, SmokeResult{
			Name:   engineName,
			Path:   goPath,
			Output: fmt.Sprintf("base=%s, categories from template", result.BaseType),
		})
		count++
	}

	return results
}
