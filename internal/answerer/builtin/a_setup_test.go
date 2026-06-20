package builtin

import (
	"github.com/seargo/seargo/internal/answerer"
)

// init sets up the GlobalAnswerer before any source-file init() functions run.
// The file name "a_setup_test.go" sorts before "random.go" alphabetically,
// ensuring this init() runs first so that random.go and statistics.go's
// init() can register with GlobalAnswerer without panicking.
func init() {
	answerer.SetGlobalAnswerer(answerer.NewAnswererStorage())
}
