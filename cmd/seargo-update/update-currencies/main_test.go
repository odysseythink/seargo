package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const mockCurrencyNamesResponse = `{
  "results": {
    "bindings": [
      {
        "iso4217": {"type": "literal", "value": "EUR"},
        "article_name": {"type": "literal", "value": "Euro", "xml:lang": "en"}
      },
      {
        "iso4217": {"type": "literal", "value": "EUR"},
        "article_name": {"type": "literal", "value": "Euro", "xml:lang": "de"}
      },
      {
        "iso4217": {"type": "literal", "value": "USD"},
        "article_name": {"type": "literal", "value": "United States dollar", "xml:lang": "en"}
      }
    ]
  }
}`

const mockCurrencyResponse = `{
  "results": {
    "bindings": [
      {
        "iso4217": {"type": "literal", "value": "EUR"},
        "label": {"type": "literal", "value": "euro", "xml:lang": "en"},
        "alias": {"type": "literal", "value": "European euro"}
      },
      {
        "iso4217": {"type": "literal", "value": "USD"},
        "label": {"type": "literal", "value": "United States dollar", "xml:lang": "en"}
      }
    ]
  }
}`

const mockRatesResponse = `{
  "base_code": "EUR",
  "time_last_update_utc": "2026-06-22 00:00:00 UTC",
  "rates": {
    "EUR": 1.0,
    "USD": 1.08
  }
}`

func TestRun_MockServer(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/sparql-results+json")
		if r.URL.Path == "/rates" {
			fmt.Fprint(w, mockRatesResponse)
			return
		}
		callCount++
		if callCount == 1 {
			fmt.Fprint(w, mockCurrencyNamesResponse)
		} else {
			fmt.Fprint(w, mockCurrencyResponse)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "currencies.json")

	if err := Run(out, srv.Client(), srv.URL, srv.URL+"/rates", false); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var db CurrencyDB
	if err := json.Unmarshal(data, &db); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	if db.Names["euro"] != "EUR" {
		t.Errorf("names[euro] = %v, want EUR", db.Names["euro"])
	}
	if !strings.Contains(fmt.Sprint(db.Names["dollar"]), "USD") {
		t.Errorf("names[dollar] = %v, want USD", db.Names["dollar"])
	}
	if db.ISO4217["EUR"]["en"] != "euro" {
		t.Errorf("iso4217[EUR][en] = %v, want euro", db.ISO4217["EUR"]["en"])
	}
	if db.Rates.Rates["USD"] <= 0 {
		t.Errorf("rates[USD] = %f, want > 0", db.Rates.Rates["USD"])
	}
	if db.Rates.Base != "EUR" {
		t.Errorf("rates.base = %q, want EUR", db.Rates.Base)
	}
}

func TestRun_SkipRates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/sparql-results+json")
		fmt.Fprint(w, mockCurrencyNamesResponse)
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "currencies.json")

	if err := Run(out, srv.Client(), srv.URL, "", true); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var db CurrencyDB
	if err := json.Unmarshal(data, &db); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if len(db.Rates.Rates) != 0 {
		t.Errorf("expected empty rates, got %v", db.Rates.Rates)
	}
}
