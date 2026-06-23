package engines

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/seargo/seargo/engines/configured"
	_ "github.com/seargo/seargo/engines/dockerhub"
	_ "github.com/seargo/seargo/engines/gentoo"
	_ "github.com/seargo/seargo/engines/githubcode"
	_ "github.com/seargo/seargo/engines/stackexchange"
	_ "github.com/seargo/seargo/engines/wikicommons"
	"github.com/seargo/seargo/internal/engine"
)

func TestP2EngineAliasesRegistered(t *testing.T) {
	aliases := []string{
		"stackoverflow", "askubuntu", "superuser",
		"wikicommons_files",
	}
	for _, name := range aliases {
		_, ok := engine.Get(name)
		assert.True(t, ok, "engine %q should be registered", name)
	}
}
