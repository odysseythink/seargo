package deps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupKM(t *testing.T) {
	entries := LookupUnit("km")
	assert.NotEmpty(t, entries)
	assert.Equal(t, "km", entries[0].Symbol)
	assert.Equal(t, "m", entries[0].SIName)
}

func TestLookupMultipleMatches(t *testing.T) {
	entries := LookupUnit("m")
	assert.NotEmpty(t, entries)
	// "m" should match the meter entry
	found := false
	for _, e := range entries {
		if e.Symbol == "m" && e.SIName == "m" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected to find the meter entry for symbol 'm'")
}

func TestLookupNotFound(t *testing.T) {
	entries := LookupUnit("nonexistent")
	assert.Empty(t, entries)
}

func TestConvertKmToMi(t *testing.T) {
	km := LookupUnit("km")
	mi := LookupUnit("mi")
	result, ok := Convert(10, km, mi)
	assert.True(t, ok)
	assert.InDelta(t, 6.21371, result, 0.001)
}

func TestConvertDifferentQuantity(t *testing.T) {
	kg := LookupUnit("kg")
	m := LookupUnit("m")
	_, ok := Convert(10, kg, m)
	assert.False(t, ok)
}

func TestConvertCelsiusToFahrenheit(t *testing.T) {
	c := LookupUnit("°c")
	f := LookupUnit("°f")
	result, ok := Convert(25, c, f)
	assert.True(t, ok)
	assert.InDelta(t, 77, result, 0.1)

	// Test freezing point
	result, ok = Convert(0, c, f)
	assert.True(t, ok)
	assert.InDelta(t, 32, result, 0.1)

	// Test boiling point
	result, ok = Convert(100, c, f)
	assert.True(t, ok)
	assert.InDelta(t, 212, result, 0.1)
}

func TestConvertInchesToCm(t *testing.T) {
	in := LookupUnit("in")
	cm := LookupUnit("cm")
	result, ok := Convert(10, in, cm)
	assert.True(t, ok)
	assert.InDelta(t, 25.4, result, 0.1)

	// Test 1 inch = 2.54 cm
	result, ok = Convert(1, in, cm)
	assert.True(t, ok)
	assert.InDelta(t, 2.54, result, 0.01)
}

func TestCaseInsensitiveLookup(t *testing.T) {
	entries := LookupUnit("KM")
	assert.NotEmpty(t, entries)
	assert.Equal(t, "km", entries[0].Symbol)
}

func TestConvertEmptySourceOrTarget(t *testing.T) {
	_, ok := Convert(10, nil, []UnitEntry{{Symbol: "m", SIName: "m"}})
	assert.False(t, ok)

	_, ok = Convert(10, []UnitEntry{{Symbol: "m", SIName: "m"}}, nil)
	assert.False(t, ok)
}

func TestLoadUnits_AppendsEntries(t *testing.T) {
	// Save and restore the global table to avoid affecting other tests.
	original := make([]UnitEntry, len(unitTable))
	copy(original, unitTable)
	t.Cleanup(func() {
		unitTable = original
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "units.json")
	data := `{
		"Q123": {"symbol": "mockm", "si_name": "Q11573", "to_si_factor": 2.5}
	}`
	require.NoError(t, os.WriteFile(path, []byte(data), 0644))

	err := LoadUnits(path)
	require.NoError(t, err)

	entries := LookupUnit("mockm")
	require.Len(t, entries, 1)
	assert.Equal(t, "mockm", entries[0].Symbol)
	assert.Equal(t, "Q11573", entries[0].SIName)
	assert.InDelta(t, 2.5, entries[0].ToSI, 1e-9)
	assert.InDelta(t, 0.4, entries[0].FromSI, 1e-9)
}

func TestLoadUnits_MissingFileReturnsError(t *testing.T) {
	err := LoadUnits(filepath.Join(t.TempDir(), "missing.json"))
	assert.Error(t, err)
}

func TestLoadUnits_MalformedFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "units.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

	err := LoadUnits(path)
	assert.Error(t, err)
}
