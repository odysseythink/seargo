package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/seargo/seargo/cmd/seargo-update/internal"
)

const defaultSPARQLURL = "https://query.wikidata.org/sparql"

const unitsSPARQL = `
SELECT DISTINCT ?item ?symbol ?tosi ?tosiUnit
WHERE
{
  ?item wdt:P31/wdt:P279 wd:Q47574 .
  ?item p:P5061 ?symbolP .
  ?symbolP ps:P5061 ?symbol ;
           wikibase:rank ?rank .
  OPTIONAL {
    ?item p:P2370 ?tosistmt .
    ?tosistmt psv:P2370 ?tosinode .
    ?tosinode wikibase:quantityAmount ?tosi .
    ?tosinode wikibase:quantityUnit ?tosiUnit .
  }
  FILTER(LANG(?symbol) = "en").
}
ORDER BY ?item DESC(?rank) ?symbol
`

// WikidataUnit matches the schema expected by SearGo's runtime.
type WikidataUnit struct {
	Symbol     string  `json:"symbol"`
	SIName     string  `json:"si_name"`
	ToSIFactor float64 `json:"to_si_factor"`
}

func main() {
	var (
		out       = flag.String("out", "data/wikidata_units.json", "output JSON path")
		sparqlURL = flag.String("sparql-url", defaultSPARQLURL, "Wikidata SPARQL endpoint")
	)
	flag.Parse()

	if err := Run(*out, nil, *sparqlURL); err != nil {
		fmt.Fprintf(os.Stderr, "update-units: %v\n", err)
		os.Exit(1)
	}
}

// Run fetches unit definitions from the Wikidata SPARQL endpoint and writes
// them to outPath. The client parameter is used by tests; nil means the
// default HTTP client.
func Run(outPath string, client fetch.Client, sparqlURL string) error {
	h := fetch.New(client)
	data := url.Values{}
	data.Set("query", unitsSPARQL)

	ctx := context.Background()
	body, err := h.PostForm(ctx, sparqlURL, data)
	if err != nil {
		return fmt.Errorf("wikidata query failed: %w", err)
	}

	var result struct {
		Results struct {
			Bindings []map[string]struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"bindings"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse SPARQL response: %w", err)
	}

	units := make(map[string]WikidataUnit)
	for _, b := range result.Results.Bindings {
		itemURL := b["item"].Value
		itemID := path.Base(itemURL)

		if _, seen := units[itemID]; seen {
			continue // keep the first binding to stay deterministic
		}

		symbol := b["symbol"].Value
		if symbol == "" {
			continue
		}

		siName := ""
		if v, ok := b["tosiUnit"]; ok && v.Value != "" {
			siName = path.Base(v.Value)
		}

		toSIFactor := 0.0
		if v, ok := b["tosi"]; ok && v.Value != "" {
			f, err := strconv.ParseFloat(v.Value, 64)
			if err == nil {
				toSIFactor = f
			}
		}

		if toSIFactor == 0 {
			continue // skip units that cannot be normalized to SI
		}

		units[itemID] = WikidataUnit{
			Symbol:     symbol,
			SIName:     siName,
			ToSIFactor: toSIFactor,
		}
	}

	if len(units) == 0 {
		return fmt.Errorf("no units fetched")
	}

	return writeJSON(outPath, units)
}

func writeJSON(outPath string, v any) error {
	keys := make([]string, 0)
	switch m := v.(type) {
	case map[string]WikidataUnit:
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		ordered := make(map[string]WikidataUnit, len(keys))
		for _, k := range keys {
			ordered[k] = m[k]
		}
		v = ordered
	}

	enc, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	enc = append(enc, '\n')

	tmp := outPath + ".tmp"
	if err := os.WriteFile(tmp, enc, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	return os.Rename(tmp, outPath)
}
