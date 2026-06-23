package upstream

import (
	"fmt"
	"reflect"
)

// Mismatch records a single difference between upstream and SearGo.
type Mismatch struct {
	Path string `json:"path"`
	Want any    `json:"want"`
	Got  any    `json:"got"`
}

// Report is the aggregate result of one or more parity checks.
type Report struct {
	Name                string             `json:"name"`
	Query               string             `json:"query"`
	Results             []NormalizedResult `json:"results,omitempty"`
	Answers             []NormalizedAnswer `json:"answers,omitempty"`
	Infoboxes           []NormalizedInfobox `json:"infoboxes,omitempty"`
	UnresponsiveEngines []string           `json:"unresponsive_engines,omitempty"`
	FailedEngines       []string           `json:"failed_engines,omitempty"`
	RedirectURL         string             `json:"redirect_url,omitempty"`
	Suppressed          []Mismatch         `json:"suppressed,omitempty"`
	Mismatches          []Mismatch         `json:"mismatches"`
}

// Diff compares two normalized responses and returns mismatches.
func Diff(name string, want, got NormalizedResponse) []Mismatch {
	var ms []Mismatch
	if want.Query != got.Query {
		ms = append(ms, Mismatch{Path: name + ".query", Want: want.Query, Got: got.Query})
	}
	if want.Total != got.Total {
		ms = append(ms, Mismatch{Path: name + ".total", Want: want.Total, Got: got.Total})
	}
	ms = append(ms, diffSlice(name+".results", want.Results, got.Results,
		func(a, b NormalizedResult) bool { return a.URL == b.URL },
		func(prefix string, a, b NormalizedResult) []Mismatch {
			var inner []Mismatch
			cmpString(&inner, prefix+".url", a.URL, b.URL)
			cmpString(&inner, prefix+".title", a.Title, b.Title)
			cmpString(&inner, prefix+".content", a.Content, b.Content)
			cmpString(&inner, prefix+".engine", a.Engine, b.Engine)
			cmpStringSlice(&inner, prefix+".engines", a.Engines, b.Engines)
			cmpString(&inner, prefix+".category", a.Category, b.Category)
			cmpString(&inner, prefix+".template", a.Template, b.Template)
			cmpString(&inner, prefix+".thumbnail_url", a.ThumbnailURL, b.ThumbnailURL)
			cmpString(&inner, prefix+".published_date", a.PublishedDate, b.PublishedDate)
			if a.Score != b.Score {
				inner = append(inner, Mismatch{Path: prefix + ".score", Want: a.Score, Got: b.Score})
			}
			for k, wantV := range a.TypedFields {
				if gotV, ok := b.TypedFields[k]; !ok || wantV != gotV {
					inner = append(inner, Mismatch{Path: prefix + ".typed_fields." + k, Want: wantV, Got: gotV})
				}
			}
			for k, gotV := range b.TypedFields {
				if _, ok := a.TypedFields[k]; !ok {
					inner = append(inner, Mismatch{Path: prefix + ".typed_fields." + k, Want: nil, Got: gotV})
				}
			}
			return inner
		})...)

	cmpStringSlice(&ms, name+".suggestions", want.Suggestions, got.Suggestions)
	cmpStringSlice(&ms, name+".corrections", want.Corrections, got.Corrections)
	ms = append(ms, diffSlice(name+".answers", want.Answers, got.Answers,
		func(a, b NormalizedAnswer) bool { return a.Answer == b.Answer },
		func(prefix string, a, b NormalizedAnswer) []Mismatch {
			var inner []Mismatch
			cmpString(&inner, prefix+".answer", a.Answer, b.Answer)
			cmpString(&inner, prefix+".url", a.URL, b.URL)
			cmpString(&inner, prefix+".content", a.Content, b.Content)
			return inner
		})...)
	ms = append(ms, diffSlice(name+".infoboxes", want.Infoboxes, got.Infoboxes,
		func(a, b NormalizedInfobox) bool { return a.Title == b.Title },
		func(prefix string, a, b NormalizedInfobox) []Mismatch {
			var inner []Mismatch
			cmpString(&inner, prefix+".title", a.Title, b.Title)
			cmpString(&inner, prefix+".url", a.URL, b.URL)
			cmpString(&inner, prefix+".content", a.Content, b.Content)
			cmpStringSlice(&inner, prefix+".engines", a.Engines, b.Engines)
			return inner
		})...)
	cmpStringSlice(&ms, name+".unresponsive_engines", want.UnresponsiveEngines, got.UnresponsiveEngines)
	return ms
}

func cmpString(ms *[]Mismatch, path, want, got string) {
	if want != got {
		*ms = append(*ms, Mismatch{Path: path, Want: want, Got: got})
	}
}

func cmpStringSlice(ms *[]Mismatch, path string, want, got []string) {
	if !reflect.DeepEqual(want, got) {
		*ms = append(*ms, Mismatch{Path: path, Want: want, Got: got})
	}
}

func diffSlice[T any](path string, want, got []T, key func(a, b T) bool, itemDiff func(prefix string, a, b T) []Mismatch) []Mismatch {
	var ms []Mismatch
	used := make([]bool, len(got))
	for _, a := range want {
		found := false
		for i, b := range got {
			if used[i] {
				continue
			}
			if key(a, b) {
				used[i] = true
				ms = append(ms, itemDiff(fmt.Sprintf("%s[%d]", path, i), a, b)...)
				found = true
				break
			}
		}
		if !found {
			ms = append(ms, Mismatch{Path: path, Want: a, Got: nil})
		}
	}
	for i, b := range got {
		if !used[i] {
			ms = append(ms, Mismatch{Path: path, Want: nil, Got: b})
		}
	}
	return ms
}
