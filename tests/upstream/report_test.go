package upstream

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReportWriter_AggregateAndWrite(t *testing.T) {
	r1 := Report{Name: "a", Query: "q1", Mismatches: []Mismatch{{Path: "a.p", Want: 1, Got: 2}}}
	r2 := Report{Name: "b", Query: "q2", Mismatches: []Mismatch{}}

	agg := AggregateReports([]Report{r1, r2})
	require.Equal(t, 2, agg.TotalCases)
	require.Equal(t, 1, agg.FailedCases)
	require.Equal(t, 1, len(agg.Mismatches))

	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")
	require.NoError(t, agg.Write(path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var decoded AggregatedReport
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, 2, decoded.TotalCases)
}
