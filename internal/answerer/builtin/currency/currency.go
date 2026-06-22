package currency

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	answerer.Register(&currencyAnswerer{})
}

// defaultCurrencyPath is the path to the generated currency data.
// It is a variable so tests can override it.
var defaultCurrencyPath = "data/currencies.json"

// currencyAnswerer provides instant currency conversion answers.
type currencyAnswerer struct{}

func (a *currencyAnswerer) Keywords() []string {
	return []string{"currency", "convert"}
}

func (a *currencyAnswerer) Info() answerer.AnswererInfo {
	return answerer.AnswererInfo{
		Name:        "currency",
		Description: "Convert amounts between currencies",
		Keywords:    a.Keywords(),
		Examples: []string{
			"currency 100 USD to EUR",
			"convert 100 dollars to euros",
		},
	}
}

var (
	currencyRegex = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([A-Za-z$€£¥]+)\s+(?:to|in)\s+([A-Za-z$€£¥]+)$`)

	dbOnce    sync.Once
	db        *currencyDB
	dbLoadErr error
)

func (a *currencyAnswerer) Answer(ctx *answerer.AnswerContext) []models.Result {
	q := stripLeadingKeyword(ctx.Query)
	if q == "" {
		return nil
	}

	matches := currencyRegex.FindStringSubmatch(q)
	if len(matches) < 4 {
		return nil
	}

	amount, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil
	}

	db := loadCurrencyDB()
	if db == nil {
		return nil
	}

	fromCode := resolveCurrencyCode(matches[2], db)
	toCode := resolveCurrencyCode(matches[3], db)
	if fromCode == "" || toCode == "" {
		return nil
	}

	converted, ok := convertCurrency(amount, fromCode, toCode, db)
	if !ok {
		return nil
	}

	resultStr := fmt.Sprintf("%s %s = %s %s",
		formatAmount(amount), fromCode,
		formatAmount(converted), toCode)

	ar := &results.AnswerResult{
		BaseResult: results.BaseResult{
			Title:   resultStr,
			Content: fmt.Sprintf("%g", converted),
			Engine:  "currency",
		},
		Answer: resultStr,
	}
	return results.ToAPIResult([]results.Result{ar})
}

// stripLeadingKeyword removes a leading "currency" or "convert" keyword from
// the query so the regex can parse the amount and currency tokens.
func stripLeadingKeyword(query string) string {
	query = strings.TrimSpace(query)
	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}
	first := strings.ToLower(words[0])
	for _, kw := range []string{"currency", "convert"} {
		if first == kw {
			return strings.TrimSpace(strings.Join(words[1:], " "))
		}
	}
	return query
}

// loadCurrencyDB loads the currency data once. If loading fails, subsequent
// calls return nil.
func loadCurrencyDB() *currencyDB {
	dbOnce.Do(func() {
		db, dbLoadErr = readCurrencyDB(defaultCurrencyPath)
	})
	if dbLoadErr != nil {
		return nil
	}
	return db
}

// resetCurrencyDBForTest forces the next call to loadCurrencyDB to reload.
// It is used by tests only.
func resetCurrencyDBForTest() {
	dbOnce = sync.Once{}
	db = nil
	dbLoadErr = nil
}

// currencyDB is the runtime view of data/currencies.json.
type currencyDB struct {
	Names   map[string]json.RawMessage   `json:"names"`
	ISO4217 map[string]map[string]string `json:"iso4217"`
	Rates   currencyRates                `json:"rates"`
}

// currencyRates stores exchange rates. The source JSON uses a flat map with
// currency codes as keys and "base"/"date" metadata mixed in, so a custom
// UnmarshalJSON is used.
type currencyRates struct {
	Base  string
	Date  string
	Table map[string]float64
}

func (r *currencyRates) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["base"]; ok {
		_ = json.Unmarshal(v, &r.Base)
	}
	if v, ok := raw["date"]; ok {
		_ = json.Unmarshal(v, &r.Date)
	}
	r.Table = make(map[string]float64)
	for k, v := range raw {
		if k == "base" || k == "date" {
			continue
		}
		var f float64
		if err := json.Unmarshal(v, &f); err == nil {
			r.Table[k] = f
		}
	}
	return nil
}

func readCurrencyDB(path string) (*currencyDB, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var db currencyDB
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

// resolveCurrencyCode converts a token (ISO code, name or symbol) into an
// ISO-4217 currency code. If the token cannot be resolved, it returns "".
func resolveCurrencyCode(token string, db *currencyDB) string {
	upper := strings.ToUpper(token)
	if _, ok := db.ISO4217[upper]; ok {
		return upper
	}

	raw, ok := db.Names[strings.ToLower(token)]
	if !ok {
		return ""
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		return strings.ToUpper(s)
	}

	var list []string
	if err := json.Unmarshal(raw, &list); err == nil && len(list) > 0 {
		return strings.ToUpper(list[0])
	}

	return ""
}

// convertCurrency converts amount from one currency to another using the
// rates table. Rates are stored as "1 base = X target" with base EUR.
func convertCurrency(amount float64, from, to string, db *currencyDB) (float64, bool) {
	fromRate, ok1 := db.Rates.Table[from]
	toRate, ok2 := db.Rates.Table[to]
	if !ok1 || !ok2 {
		return 0, false
	}
	if fromRate == 0 {
		return 0, false
	}
	return amount / fromRate * toRate, true
}

// formatAmount formats a numeric amount for display, keeping two decimal
// places for fractional values and none for whole numbers.
func formatAmount(v float64) string {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return strconv.FormatFloat(v, 'f', 2, 64)
}
