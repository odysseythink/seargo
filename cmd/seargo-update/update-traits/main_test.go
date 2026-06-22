package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/seargo/seargo/cmd/seargo-update/internal"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
)

func TestRun_MockFetcher(t *testing.T) {
	registerTraitFetcher("mockengine", func(ctx context.Context, h *fetch.Helper, cfg config.EngineConfig) (engine.EngineTraits, error) {
		return engine.EngineTraits{
			DataType:  "traits_v1",
			Languages: map[string]string{"en": "en-US"},
			Regions:   map[string]string{"us": "en-US"},
			AllLocale: "all",
		}, nil
	})

	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.yml")
	outPath := filepath.Join(dir, "engine_traits.json")

	settings := `engines:
  - name: mockengine
    engine: mockengine
    inactive: false
`
	if err := os.WriteFile(configPath, []byte(settings), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := Run(outPath, nil, configPath, nil); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var tm engine.EngineTraitsMap
	if err := json.Unmarshal(data, &tm); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	traits, ok := tm["mockengine"]
	if !ok {
		t.Fatal("mockengine traits not found")
	}
	if traits.DataType != "traits_v1" {
		t.Errorf("data_type = %q, want traits_v1", traits.DataType)
	}
	if traits.Languages["en"] != "en-US" {
		t.Errorf("languages[en] = %q, want en-US", traits.Languages["en"])
	}
}

func TestRun_IncrementalMerge(t *testing.T) {
	registerTraitFetcher("newengine", func(ctx context.Context, h *fetch.Helper, cfg config.EngineConfig) (engine.EngineTraits, error) {
		return engine.EngineTraits{DataType: "traits_v1", Languages: map[string]string{"de": "de"}}, nil
	})

	dir := t.TempDir()
	configPath := filepath.Join(dir, "settings.yml")
	outPath := filepath.Join(dir, "engine_traits.json")

	existing := `{"oldengine":{"data_type":"traits_v1","languages":{"fr":"fr"},"regions":{},"all_locale":""}}`
	if err := os.WriteFile(outPath, []byte(existing), 0644); err != nil {
		t.Fatalf("write existing traits: %v", err)
	}

	settings := `engines:
  - name: oldengine
    engine: oldengine
    inactive: false
  - name: newengine
    engine: newengine
    inactive: false
`
	if err := os.WriteFile(configPath, []byte(settings), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := Run(outPath, nil, configPath, []string{"newengine"}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var tm engine.EngineTraitsMap
	if err := json.Unmarshal(data, &tm); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if _, ok := tm["oldengine"]; !ok {
		t.Error("existing oldengine traits should be preserved")
	}
	if _, ok := tm["newengine"]; !ok {
		t.Error("newengine traits should be added")
	}
}
