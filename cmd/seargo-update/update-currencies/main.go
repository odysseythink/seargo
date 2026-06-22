package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/seargo/seargo/cmd/seargo-update/internal"
)

const (
	defaultSPARQLURL = "https://query.wikidata.org/sparql"
	defaultRatesURL  = "https://open.er-api.com/v6/latest/EUR"
)

const currencyNamesSPARQL = `
SELECT DISTINCT ?iso4217 ?article_name WHERE {
  ?item wdt:P498 ?iso4217 .
  ?article schema:about ?item ;
           schema:name ?article_name ;
           schema:isPartOf [ wikibase:wikiGroup "wikipedia" ]
  MINUS { ?item wdt:P582 ?end_data . }
  MINUS { ?item wdt:P31/wdt:P279* wd:Q15893266 . }
  FILTER(LANG(?article_name) IN (%LANGUAGES_SPARQL%)).
}
ORDER BY ?iso4217 ?article_name
`

const currencySPARQL = `
SELECT DISTINCT ?iso4217 ?unit ?unicode ?label ?alias WHERE {
  ?item wdt:P498 ?iso4217; rdfs:label ?label.
  OPTIONAL { ?item skos:altLabel ?alias FILTER (LANG (?alias) = LANG(?label)). }
  OPTIONAL { ?item wdt:P5061 ?unit. }
  OPTIONAL { ?item wdt:P489 ?symbol.
             ?symbol wdt:P487 ?unicode. }
  MINUS { ?item wdt:P582 ?end_data . }
  MINUS { ?item wdt:P31/wdt:P279* wd:Q15893266 . }
  FILTER(LANG(?label) IN (%LANGUAGES_SPARQL%)).
}
ORDER BY ?iso4217 ?unit ?unicode ?label ?alias
`

// Languages used for Wikidata currency labels.
var languages = []string{
	"en", "de", "es", "fr", "it", "ja", "nl", "pl", "pt", "ru", "zh",
	"ar", "bg", "ca", "cs", "da", "el", "fi", "he", "hi", "hr", "hu",
	"id", "ko", "ms", "no", "ro", "sk", "sl", "sv", "th", "tr", "uk",
	"vi",
}

// CurrencyDB is the output schema for data/currencies.json.
type CurrencyDB struct {
	Names   map[string]any               `json:"names"`
	ISO4217 map[string]map[string]string `json:"iso4217"`
	Rates   RateTable                    `json:"rates"`
}

// RateTable holds exchange rates against a base currency. The JSON
// representation is flat: {"base":"EUR","date":"...","EUR":1.0,"USD":1.08}.
type RateTable struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"-"`
}

// MarshalJSON flattens the rates map alongside base and date.
func (rt RateTable) MarshalJSON() ([]byte, error) {
	flat := make(map[string]interface{}, len(rt.Rates)+2)
	flat["base"] = rt.Base
	flat["date"] = rt.Date
	for k, v := range rt.Rates {
		flat[k] = v
	}
	return json.Marshal(flat)
}

// UnmarshalJSON reconstructs the rates map from the flat JSON object.
func (rt *RateTable) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	rt.Base, _ = raw["base"].(string)
	rt.Date, _ = raw["date"].(string)
	rt.Rates = make(map[string]float64)
	for k, v := range raw {
		if k == "base" || k == "date" {
			continue
		}
		switch n := v.(type) {
		case float64:
			rt.Rates[k] = n
		case json.Number:
			f, _ := n.Float64()
			rt.Rates[k] = f
		}
	}
	return nil
}

func main() {
	var (
		out       = flag.String("out", "data/currencies.json", "output JSON path")
		sparqlURL = flag.String("sparql-url", defaultSPARQLURL, "Wikidata SPARQL endpoint")
		ratesURL  = flag.String("rates-url", defaultRatesURL, "exchange-rates API URL")
		skipRates = flag.Bool("skip-rates", false, "skip fetching exchange rates")
	)
	flag.Parse()

	if err := Run(*out, nil, *sparqlURL, *ratesURL, *skipRates); err != nil {
		fmt.Fprintf(os.Stderr, "update-currencies: %v\n", err)
		os.Exit(1)
	}
}

// Run fetches currency names/labels from Wikidata and exchange rates from a
// public API, then writes data/currencies.json.
func Run(outPath string, client fetch.Client, sparqlURL, ratesURL string, skipRates bool) error {
	db := CurrencyDB{
		Names:   make(map[string]any),
		ISO4217: make(map[string]map[string]string),
		Rates: RateTable{
			Base:  "EUR",
			Date:  time.Now().UTC().Format("2006-01-02"),
			Rates: make(map[string]float64),
		},
	}

	h := fetch.New(client)
	ctx := context.Background()

	langs := languagesSPARQL()

	// 1. Wikipedia article names.
	if err := queryCurrencies(ctx, h, sparqlURL, strings.ReplaceAll(currencyNamesSPARQL, "%LANGUAGES_SPARQL%", langs), func(b sparqlBinding) {
		iso := b["iso4217"].Value
		name := b["article_name"].Value
		lang := b["article_name"].Lang
		addName(&db, name, iso)
		addLabel(&db, name, iso, lang)
	}); err != nil {
		return err
	}

	// 2. Wikidata labels/aliases/symbols.
	if err := queryCurrencies(ctx, h, sparqlURL, strings.ReplaceAll(currencySPARQL, "%LANGUAGES_SPARQL%", langs), func(b sparqlBinding) {
		iso := b["iso4217"].Value
		if v, ok := b["label"]; ok {
			addName(&db, v.Value, iso)
			addLabel(&db, v.Value, iso, v.Lang)
		}
		if v, ok := b["alias"]; ok {
			addName(&db, v.Value, iso)
		}
		if v, ok := b["unicode"]; ok {
			addNameRaw(&db, v.Value, iso)
		}
		if v, ok := b["unit"]; ok {
			addNameRaw(&db, v.Value, iso)
		}
	}); err != nil {
		return err
	}

	// 3. Static fallbacks.
	addName(&db, "euro", "EUR")
	addName(&db, "euros", "EUR")
	addName(&db, "dollar", "USD")
	addName(&db, "dollars", "USD")
	addName(&db, "peso", "MXN")
	addName(&db, "pesos", "MXN")

	// 4. Normalize single-element lists to scalars.
	for name, v := range db.Names {
		if list, ok := v.([]string); ok && len(list) == 1 {
			db.Names[name] = list[0]
		}
	}

	// 5. Exchange rates.
	if !skipRates {
		rates, err := fetchRates(ctx, h, ratesURL)
		if err != nil {
			return fmt.Errorf("fetch rates: %w", err)
		}
		db.Rates = rates
	}

	return writeJSON(outPath, db)
}

type sparqlBinding map[string]struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Lang  string `json:"xml:lang"`
}

func queryCurrencies(ctx context.Context, h *fetch.Helper, sparqlURL, query string, fn func(sparqlBinding)) error {
	data := url.Values{}
	data.Set("query", query)
	body, err := h.PostForm(ctx, sparqlURL, data)
	if err != nil {
		return fmt.Errorf("wikidata query failed: %w", err)
	}

	var result struct {
		Results struct {
			Bindings []sparqlBinding `json:"bindings"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse SPARQL response: %w", err)
	}

	for _, b := range result.Results.Bindings {
		fn(b)
	}
	return nil
}

func languagesSPARQL() string {
	parts := make([]string, len(languages))
	for i, l := range languages {
		parts[i] = fmt.Sprintf("%q", l)
	}
	return strings.Join(parts, ", ")
}

var spaceRE = regexp.MustCompile(` +`)

func normalizeName(name string) string {
	s := norm.NFKD.String(strings.ToLower(name))
	s = strings.ReplaceAll(s, "-", " ")
	s = spaceRE.ReplaceAllString(s, " ")
	for _, sep := range []string{"(", ":"} {
		if idx := strings.Index(s, sep); idx >= 0 {
			s = strings.TrimSpace(s[:idx])
		}
	}
	return s
}

func addName(db *CurrencyDB, name, iso4217 string) {
	addNameRaw(db, normalizeName(name), iso4217)
}

func addNameRaw(db *CurrencyDB, name, iso4217 string) {
	if name == "" {
		return
	}
	existing, ok := db.Names[name]
	if !ok {
		db.Names[name] = iso4217
		return
	}
	if existing == iso4217 {
		return
	}

	var list []string
	switch v := existing.(type) {
	case string:
		list = []string{v}
	case []string:
		list = v
	}

	for _, item := range list {
		if item == iso4217 {
			return
		}
	}
	list = append(list, iso4217)
	db.Names[name] = list
}

func addLabel(db *CurrencyDB, label, iso4217, language string) {
	if label == "" || language == "" {
		return
	}
	labels, ok := db.ISO4217[iso4217]
	if !ok {
		labels = make(map[string]string)
		db.ISO4217[iso4217] = labels
	}
	// The Wikidata label overwrites the article name so that lowercase forms
	// ("euro") take precedence over title-case article names ("Euro").
	labels[language] = label
}

func fetchRates(ctx context.Context, h *fetch.Helper, ratesURL string) (RateTable, error) {
	body, err := h.Get(ctx, ratesURL)
	if err != nil {
		return RateTable{}, err
	}

	var result struct {
		Base  string             `json:"base_code"`
		Date  string             `json:"time_last_update_utc"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return RateTable{}, fmt.Errorf("parse rates response: %w", err)
	}

	// Fallback for APIs that use different field names.
	if result.Base == "" {
		var alt struct {
			Base  string             `json:"base"`
			Date  string             `json:"date"`
			Rates map[string]float64 `json:"rates"`
		}
		if err := json.Unmarshal(body, &alt); err == nil && alt.Base != "" {
			result.Base = alt.Base
			result.Date = alt.Date
			result.Rates = alt.Rates
		}
	}

	rt := RateTable{
		Base:  result.Base,
		Rates: result.Rates,
	}
	if result.Date != "" {
		// Try to parse common date formats.
		for _, layout := range []string{"2006-01-02", time.RFC1123, time.RFC1123Z, time.RFC3339} {
			if d, err := time.Parse(layout, result.Date); err == nil {
				rt.Date = d.Format("2006-01-02")
				break
			}
		}
	}
	if rt.Date == "" {
		rt.Date = time.Now().UTC().Format("2006-01-02")
	}
	return rt, nil
}

func writeJSON(outPath string, db CurrencyDB) error {
	// Sort map keys by building ordered maps for encoding.
	enc, err := json.MarshalIndent(toOrdered(db), "", "    ")
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

func toOrdered(db CurrencyDB) map[string]any {
	out := make(map[string]any)
	out["names"] = sortedMap(db.Names)

	isoOrdered := make(map[string]any)
	for iso, labels := range db.ISO4217 {
		isoOrdered[iso] = sortedStringMap(labels)
	}
	out["iso4217"] = sortedMap(isoOrdered)

	out["rates"] = db.Rates
	return out
}

func sortedMap(m map[string]any) map[string]any {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]any, len(keys))
	for _, k := range keys {
		ordered[k] = m[k]
	}
	return ordered
}

func sortedStringMap(m map[string]string) map[string]string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]string, len(keys))
	for _, k := range keys {
		ordered[k] = m[k]
	}
	return ordered
}
