package deps

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
