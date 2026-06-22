package weather

import (
	"fmt"
	"strings"
	"time"
)

// WeatherData holds the current weather for a location.
type WeatherData struct {
	Location      GeoLocation
	Condition     string
	Temperature   Temperature
	FeelsLike     Temperature
	WindSpeed     WindSpeed
	WindDirection Compass
	Pressure      Pressure
	Humidity      RelativeHumidity
	CloudCover    float64
	IsDay         bool
	Forecasts     []Forecast
}

// Forecast holds a weather forecast at a point in time.
type Forecast struct {
	DateTime time.Time
	WeatherData
}

// GeoLocation holds the resolved location from a geocoding service.
type GeoLocation struct {
	Name        string
	Latitude    float64
	Longitude   float64
	Elevation   float64
	CountryCode string
	Timezone    string
}

// Temperature is a temperature value with its native unit.
type Temperature struct {
	Value float64
	Unit  string
}

// WindSpeed is a wind speed value with its native unit.
type WindSpeed struct {
	Value float64
	Unit  string
}

// Compass is a directional value with its native unit.
type Compass struct {
	Value float64
	Unit  string
}

// Pressure is an atmospheric pressure value with its native unit.
type Pressure struct {
	Value float64
	Unit  string
}

// RelativeHumidity is a humidity percentage value.
type RelativeHumidity struct {
	Value float64
}

// L10n formats the temperature for the given locale, converting to °F for US locales.
func (t Temperature) L10n(locale string) string {
	if isUSLocale(locale) {
		return fmt.Sprintf("%.1f °F", celsiusToFahrenheit(t.Value))
	}
	return fmt.Sprintf("%.1f °C", t.Value)
}

// L10n formats the wind speed with its unit.
func (w WindSpeed) L10n(_ string) string {
	return fmt.Sprintf("%.1f %s", w.Value, w.Unit)
}

// L10n formats the compass direction with its unit.
func (c Compass) L10n(_ string) string {
	return fmt.Sprintf("%.0f%s", c.Value, c.Unit)
}

// L10n formats the pressure with its unit.
func (p Pressure) L10n(_ string) string {
	return fmt.Sprintf("%.1f %s", p.Value, p.Unit)
}

// L10n formats the relative humidity as a percentage.
func (h RelativeHumidity) L10n(_ string) string {
	return fmt.Sprintf("%.0f%%", h.Value)
}

func isUSLocale(locale string) bool {
	locale = strings.ToLower(locale)
	if locale == "en-us" {
		return true
	}
	parts := strings.Split(locale, "-")
	if len(parts) > 0 && parts[len(parts)-1] == "us" {
		return true
	}
	return false
}

func celsiusToFahrenheit(c float64) float64 {
	return c*9.0/5.0 + 32.0
}
