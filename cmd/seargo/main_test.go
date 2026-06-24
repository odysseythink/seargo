package main

import (
	"testing"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineImports_P2EnginesRegistered(t *testing.T) {
	expected := []string{
		"docker_hub",
		"gentoo",
		"github_code",
		"hoogle",
		"mdn",
		"mankier",
		"openairedatasets",
		"openairepublications",
		"stackoverflow",
		"askubuntu",
		"superuser",
		"wikicommons",
	}
	for _, name := range expected {
		_, ok := engine.Get(name)
		assert.Truef(t, ok, "engine %q should be registered", name)
	}
}

func TestConfig_P2CategoryTabsPopulated(t *testing.T) {
	cfg, err := config.Load("../../configs/settings.yml")
	require.NoError(t, err)

	assert.Subset(t, cfg.CategoriesAsTabs["it"].Engines, []string{
		"docker_hub", "hoogle", "mdn", "mankier", "stackoverflow", "askubuntu", "superuser", "gentoo", "github_code",
	})
	assert.Subset(t, cfg.CategoriesAsTabs["science"].Engines, []string{
		"openairedatasets", "openairepublications",
	})
	assert.Subset(t, cfg.CategoriesAsTabs["files"].Engines, []string{
		"wikicommons",
	})
}
