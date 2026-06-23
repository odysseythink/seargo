package upstream

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// GlobalReports is the collector used by the upstream test run.
var GlobalReports ReportCollector

// ReportCollector gathers reports from all tests safely.
type ReportCollector struct {
	mu      sync.Mutex
	reports []Report
}

// Record adds a report.
func (rc *ReportCollector) Record(r Report) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.reports = append(rc.reports, r)
}

// Reports returns a copy of the collected reports.
func (rc *ReportCollector) Reports() []Report {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	out := make([]Report, len(rc.reports))
	copy(out, rc.reports)
	return out
}

// AggregatedReport is the machine-readable summary of a full run.
type AggregatedReport struct {
	GeneratedAt  string     `json:"generatedAt"`
	DurationMs   int64      `json:"durationMs"`
	TotalCases   int        `json:"totalCases"`
	FailedCases  int        `json:"failedCases"`
	SkippedCases int        `json:"skippedCases"`
	Reports      []Report   `json:"reports"`
	Mismatches   []Mismatch `json:"mismatches"`
}

// AggregateReports builds a summary from individual reports.
func AggregateReports(reports []Report) AggregatedReport {
	agg := AggregatedReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		TotalCases:  len(reports),
		Reports:     reports,
	}
	for _, r := range reports {
		if len(r.Mismatches) > 0 {
			agg.FailedCases++
			agg.Mismatches = append(agg.Mismatches, r.Mismatches...)
		}
	}
	return agg
}

// Write serializes the aggregated report to path, creating parent dirs if needed.
func (a AggregatedReport) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
