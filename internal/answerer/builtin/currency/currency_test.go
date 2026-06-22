package currency

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeMockCurrencyDB(t *testing.T, data string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "currencies.json")
	require.NoError(t, os.WriteFile(path, []byte(data), 0644))
	return path
}

func setupWithMock(t *testing.T, data string) *currencyAnswerer {
	t.Helper()
	oldPath := defaultCurrencyPath
	defaultCurrencyPath = writeMockCurrencyDB(t, data)
	resetCurrencyDBForTest()
	t.Cleanup(func() {
		defaultCurrencyPath = oldPath
		resetCurrencyDBForTest()
	})
	return &currencyAnswerer{}
}

const mockCurrencyDB = `{
  "iso4217": {
    "USD": {"en": "United States dollar"},
    "EUR": {"en": "euro"},
    "GBP": {"en": "British pound"},
    "XYZ": {"en": "unknown currency"}
  },
  "names": {
    "usd": "USD",
    "eur": "EUR",
    "$": ["USD", "CAD"],
    "dollar": "USD",
    "dollars": "USD",
    "euro": "EUR",
    "euros": "EUR",
    "pound": "GBP"
  },
  "rates": {
    "base": "EUR",
    "date": "2026-06-22",
    "EUR": 1.0,
    "USD": 1.146817,
    "GBP": 0.867207
  }
}`

func TestCurrencyAnswerer_USDToEUR(t *testing.T) {
	a := setupWithMock(t, mockCurrencyDB)
	res := a.Answer(&answerer.AnswerContext{Query: "currency 100 USD to EUR"})
	require.Len(t, res, 1)
	assert.Equal(t, "answer", res[0].Kind)
	assert.Equal(t, "currency", res[0].Engine)
	assert.Contains(t, res[0].Title, "100 USD")
	assert.Contains(t, res[0].Title, "EUR")
	assert.Contains(t, res[0].Title, "=")
}

func TestCurrencyAnswerer_NamesToCodes(t *testing.T) {
	a := setupWithMock(t, mockCurrencyDB)
	res := a.Answer(&answerer.AnswerContext{Query: "convert 100 dollars to euros"})
	require.Len(t, res, 1)
	assert.Equal(t, "answer", res[0].Kind)
	assert.Contains(t, res[0].Title, "USD")
	assert.Contains(t, res[0].Title, "EUR")
}

func TestCurrencyAnswerer_UnknownCurrency(t *testing.T) {
	a := setupWithMock(t, mockCurrencyDB)
	res := a.Answer(&answerer.AnswerContext{Query: "currency 100 ABC to EUR"})
	assert.Empty(t, res)
}

func TestCurrencyAnswerer_MissingRate(t *testing.T) {
	a := setupWithMock(t, mockCurrencyDB)
	res := a.Answer(&answerer.AnswerContext{Query: "currency 100 XYZ to EUR"})
	assert.Empty(t, res)
}

func TestCurrencyAnswerer_NoKeywordStrip(t *testing.T) {
	a := setupWithMock(t, mockCurrencyDB)
	res := a.Answer(&answerer.AnswerContext{Query: "currency100 USD to EUR"})
	assert.Empty(t, res)
}

func TestCurrencyAnswerer_MissingFile(t *testing.T) {
	oldPath := defaultCurrencyPath
	defaultCurrencyPath = filepath.Join(t.TempDir(), "missing.json")
	resetCurrencyDBForTest()
	t.Cleanup(func() {
		defaultCurrencyPath = oldPath
		resetCurrencyDBForTest()
	})

	a := &currencyAnswerer{}
	res := a.Answer(&answerer.AnswerContext{Query: "currency 100 USD to EUR"})
	assert.Empty(t, res)
}
