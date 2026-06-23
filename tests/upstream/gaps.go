package upstream

import (
	"encoding/json"
	"os"
	"strings"
)

// GapRule describes a known mismatch that should not fail the suite.
type GapRule struct {
	Name       string `json:"name"`
	PathPrefix string `json:"pathPrefix"`
	Reason     string `json:"reason"`
}

// GapRules is a list of suppression rules.
type GapRules []GapRule

// LoadGaps reads gap rules from a JSON file.
func LoadGaps(path string) (GapRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rules GapRules
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// Filter removes mismatches that match a documented gap for this report.
func (g GapRules) Filter(r Report) Report {
	out := r
	out.Mismatches = nil
	out.Suppressed = nil
	for _, m := range r.Mismatches {
		if g.matches(r.Name, m.Path) {
			out.Suppressed = append(out.Suppressed, m)
		} else {
			out.Mismatches = append(out.Mismatches, m)
		}
	}
	return out
}

func (g GapRules) matches(reportName, path string) bool {
	for _, rule := range g {
		if rule.Name != "" && !strings.HasPrefix(reportName, rule.Name) {
			continue
		}
		if strings.HasPrefix(path, rule.PathPrefix) {
			return true
		}
	}
	return false
}
