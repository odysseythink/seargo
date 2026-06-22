package bases

import (
	"flag"
	"os"
	"testing"
)

// TestMain redirects mlog output to stderr so that tests do not trigger the
// file-sink path, which can abort the process when multiple severity levels
// attempt to create log files with the same name within the same second
// (mlog upstream issue: logName does not include the severity tag).
func TestMain(m *testing.M) {
	flag.Parse()
	_ = flag.Set("logtostderr", "true")
	code := m.Run()
	os.Exit(code)
}
