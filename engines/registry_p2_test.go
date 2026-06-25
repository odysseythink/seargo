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

func TestP2DedicatedEnginesRegistered(t *testing.T) {
	// Dedicated engine implementations register their engine type name.
	// Multi-instance engines (stackoverflow/askubuntu/superuser) are
	// resolved from the "stackexchange" engine via EngineType in the loader.
	engines := []string{
		"stackexchange",
		"wikicommons",
	}
	for _, name := range engines {
		_, ok := engine.Get(name)
		assert.True(t, ok, "engine %q should be registered", name)
	}
}
