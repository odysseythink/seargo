package deps

import (
	"math"
	"strings"
)

// UnitEntry describes a unit and how to convert it to/from its SI base.
type UnitEntry struct {
	Symbol string
	SIName string
	ToSI   float64
	FromSI float64
}

var unitTable = []UnitEntry{
	// Length (SI: m)
	{Symbol: "m", SIName: "m", ToSI: 1, FromSI: 1},
	{Symbol: "km", SIName: "m", ToSI: 1000, FromSI: 0.001},
	{Symbol: "cm", SIName: "m", ToSI: 0.01, FromSI: 100},
	{Symbol: "mm", SIName: "m", ToSI: 0.001, FromSI: 1000},
	{Symbol: "mi", SIName: "m", ToSI: 1609.344, FromSI: 0.000621371},
	{Symbol: "in", SIName: "m", ToSI: 0.0254, FromSI: 39.3701},
	{Symbol: "ft", SIName: "m", ToSI: 0.3048, FromSI: 3.28084},
	{Symbol: "yd", SIName: "m", ToSI: 0.9144, FromSI: 1.09361},
	{Symbol: "nm", SIName: "m", ToSI: 1852, FromSI: 0.000539957},
	// Mass (SI: kg)
	{Symbol: "kg", SIName: "kg", ToSI: 1, FromSI: 1},
	{Symbol: "g", SIName: "kg", ToSI: 0.001, FromSI: 1000},
	{Symbol: "mg", SIName: "kg", ToSI: 0.000001, FromSI: 1e6},
	{Symbol: "lb", SIName: "kg", ToSI: 0.453592, FromSI: 2.20462},
	{Symbol: "oz", SIName: "kg", ToSI: 0.0283495, FromSI: 35.274},
	{Symbol: "t", SIName: "kg", ToSI: 1000, FromSI: 0.001},
	// Temperature (non-linear, special-cased in Convert)
	{Symbol: "°c", SIName: "°c", ToSI: 1, FromSI: 1},
	{Symbol: "°f", SIName: "°f", ToSI: 1, FromSI: 1},
	{Symbol: "k", SIName: "k", ToSI: 1, FromSI: 1},
	// Time (SI: s)
	{Symbol: "s", SIName: "s", ToSI: 1, FromSI: 1},
	{Symbol: "min", SIName: "s", ToSI: 60, FromSI: 0.0166667},
	{Symbol: "h", SIName: "s", ToSI: 3600, FromSI: 0.000277778},
	{Symbol: "d", SIName: "s", ToSI: 86400, FromSI: 0.0000115741},
	// Speed (SI: km/h — treated as its own quantity class)
	{Symbol: "km/h", SIName: "km/h", ToSI: 1, FromSI: 1},
	{Symbol: "mph", SIName: "km/h", ToSI: 1.60934, FromSI: 0.621371},
	// Area (SI: m²)
	{Symbol: "m²", SIName: "m²", ToSI: 1, FromSI: 1},
	{Symbol: "km²", SIName: "m²", ToSI: 1e6, FromSI: 1e-6},
	{Symbol: "ha", SIName: "m²", ToSI: 10000, FromSI: 0.0001},
	{Symbol: "ac", SIName: "m²", ToSI: 4046.86, FromSI: 0.000247105},
	// Volume (SI: l)
	{Symbol: "l", SIName: "l", ToSI: 1, FromSI: 1},
	{Symbol: "ml", SIName: "l", ToSI: 0.001, FromSI: 1000},
	{Symbol: "gal", SIName: "l", ToSI: 3.78541, FromSI: 0.264172},
}

// LookupUnit searches the unit table by symbol (case-insensitive).
// Returns all matching entries (there may be multiple for symbols like "m").
func LookupUnit(symbol string) []UnitEntry {
	var results []UnitEntry
	for _, u := range unitTable {
		if strings.EqualFold(u.Symbol, symbol) {
			results = append(results, u)
		}
	}
	return results
}

// Convert converts a value from source units to target units.
// source and target are the results of LookupUnit calls.
// It returns false if conversion is not possible (different quantity classes
// or empty input). Temperature conversions are special-cased.
func Convert(value float64, source, target []UnitEntry) (float64, bool) {
	if len(source) == 0 || len(target) == 0 {
		return 0, false
	}
	s, t := source[0], target[0]

	// Temperature special case
	if isTempUnit(s.SIName) && isTempUnit(t.SIName) {
		return convertTemperature(value, s.SIName, t.SIName), true
	}

	// Different quantity — reject
	if s.SIName != t.SIName {
		return 0, false
	}

	// Linear: value -> SI base -> target
	return value * s.ToSI * t.FromSI, true
}

func isTempUnit(siName string) bool {
	return siName == "°c" || siName == "°f" || siName == "k"
}

// convertTemperature converts between Celsius, Fahrenheit and Kelvin.
func convertTemperature(value float64, from, to string) float64 {
	// Convert to Celsius first
	var inC float64
	switch from {
	case "°c":
		inC = value
	case "°f":
		inC = (value - 32) * 5 / 9
	case "k":
		inC = value - 273.15
	default:
		return value
	}

	// Convert from Celsius to target
	switch to {
	case "°c":
		return inC
	case "°f":
		return inC*9/5 + 32
	case "k":
		return inC + 273.15
	default:
		return math.NaN()
	}
}
