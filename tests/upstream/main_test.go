//go:build upstream

package upstream

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// findModuleRoot walks up from dir until it finds a directory containing go.mod.
func findModuleRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

func TestMain(m *testing.M) {
	start := time.Now()
	code := m.Run()
	agg := AggregateReports(GlobalReports.Reports())
	agg.DurationMs = time.Since(start).Milliseconds()

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	root := findModuleRoot(cwd)
	filename := fmt.Sprintf("upstream-parity-%s.json", time.Now().UTC().Format("20060102-150405"))
	path := filepath.Join(root, ".ody-code", "test-reports", filename)
	_ = agg.Write(path)
	os.Exit(code)
}
