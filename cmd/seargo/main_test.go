package main

import (
	"testing"

	"github.com/seargo/seargo/internal/engine"
	"github.com/stretchr/testify/assert"
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
