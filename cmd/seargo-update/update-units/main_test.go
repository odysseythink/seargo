package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const mockUnitsResponse = `{
  "results": {
    "bindings": [
      {
        "item": {"type": "uri", "value": "http://www.wikidata.org/entity/Q11573"},
        "symbol": {"type": "literal", "value": "m"},
        "tosi": {"type": "literal", "value": "1.0"},
        "tosiUnit": {"type": "uri", "value": "http://www.wikidata.org/entity/Q11573"}
      },
      {
        "item": {"type": "uri", "value": "http://www.wikidata.org/entity/Q11579"},
        "symbol": {"type": "literal", "value": "K"},
        "tosi": {"type": "literal", "value": "1.0"},
        "tosiUnit": {"type": "uri", "value": "http://www.wikidata.org/entity/Q11579"}
      },
      {
        "item": {"type": "uri", "value": "http://www.wikidata.org/entity/Q253276"},
        "symbol": {"type": "literal", "value": "mi"},
        "tosi": {"type": "literal", "value": "1609.344"},
        "tosiUnit": {"type": "uri", "value": "http://www.wikidata.org/entity/Q11573"}
      }
    ]
  }
}`

func TestRun_MockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("expected form content type, got %s", ct)
		}
		w.Header().Set("Content-Type", "application/sparql-results+json")
		fmt.Fprint(w, mockUnitsResponse)
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "wikidata_units.json")

	if err := Run(out, srv.Client(), srv.URL); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var units map[string]struct {
		Symbol     string  `json:"symbol"`
		SIName     string  `json:"si_name"`
		ToSIFactor float64 `json:"to_si_factor"`
	}
	if err := jsonUnmarshal(data, &units); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	for _, id := range []string{"Q11573", "Q11579"} {
		u, ok := units[id]
		if !ok {
			t.Fatalf("missing unit %s", id)
		}
		if u.ToSIFactor <= 0 {
			t.Errorf("unit %s to_si_factor = %f, want > 0", id, u.ToSIFactor)
		}
	}

	mi, ok := units["Q253276"]
	if !ok {
		t.Fatal("missing Q253276 (mile)")
	}
	if mi.Symbol != "mi" {
		t.Errorf("Q253276 symbol = %q, want mi", mi.Symbol)
	}
	if mi.SIName != "Q11573" {
		t.Errorf("Q253276 si_name = %q, want Q11573", mi.SIName)
	}
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
