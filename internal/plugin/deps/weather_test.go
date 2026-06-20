package deps

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTimeFormat(t *testing.T) {
	now := time.Date(2026, 6, 20, 15, 30, 0, 0, time.UTC)
	dt := NewDateTime(now, "UTC")
	formatted := dt.Format()
	assert.Contains(t, formatted, "2026-06-20 15:30")
	assert.Contains(t, formatted, "UTC")
}

func TestDateTimeLocaleFormat(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	now := time.Date(2026, 6, 20, 15, 30, 0, 0, loc)
	dt := NewDateTime(now, "Europe/Berlin")
	formatted := dt.Format()
	assert.Contains(t, formatted, "2026-06-20 15:30")
	assert.Contains(t, formatted, "Europe/Berlin")
}

func TestKnownCities(t *testing.T) {
	loc, ok := GeoLocationByQuery("Berlin")
	assert.True(t, ok)
	assert.Equal(t, "Berlin", loc.Name)
	assert.Equal(t, "Europe/Berlin", loc.Timezone)

	loc, ok = GeoLocationByQuery("Tokyo")
	assert.True(t, ok)
	assert.Equal(t, "Tokyo", loc.Name)
	assert.Equal(t, "Asia/Tokyo", loc.Timezone)

	loc, ok = GeoLocationByQuery("San Francisco")
	assert.True(t, ok)
	assert.Equal(t, "San Francisco", loc.Name)
	assert.Equal(t, "America/Los_Angeles", loc.Timezone)
}

func TestUnknownCity(t *testing.T) {
	_, ok := GeoLocationByQuery("Atlantis")
	assert.False(t, ok)
}

func TestCaseInsensitive(t *testing.T) {
	loc, ok := GeoLocationByQuery("berlin")
	assert.True(t, ok)
	assert.Equal(t, "Berlin", loc.Name)

	loc, ok = GeoLocationByQuery("NEW YORK")
	assert.True(t, ok)
	assert.Equal(t, "New York", loc.Name)

	loc, ok = GeoLocationByQuery("sAn FrAnCiScO")
	assert.True(t, ok)
	assert.Equal(t, "San Francisco", loc.Name)
}
