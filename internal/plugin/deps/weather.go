package deps

import (
	"strings"
	"time"
)

// GeoLocation represents a city with its timezone information.
type GeoLocation struct {
	Name     string
	Timezone string
}

// DateTime holds a time and a timezone string and can format them together.
type DateTime struct {
	t  time.Time
	tz string
}

// NewDateTime creates a new DateTime with the given time and timezone.
func NewDateTime(t time.Time, tz string) DateTime {
	return DateTime{t: t, tz: tz}
}

// Format returns the date-time string in "2006-01-02 15:04 TZ" format.
func (d DateTime) Format() string {
	loc, err := time.LoadLocation(d.tz)
	if err != nil {
		return d.t.Format("2006-01-02 15:04") + " " + d.tz
	}
	return d.t.In(loc).Format("2006-01-02 15:04") + " " + loc.String()
}

var geoLocations = []GeoLocation{
	{Name: "Berlin", Timezone: "Europe/Berlin"},
	{Name: "London", Timezone: "Europe/London"},
	{Name: "Paris", Timezone: "Europe/Paris"},
	{Name: "Tokyo", Timezone: "Asia/Tokyo"},
	{Name: "New York", Timezone: "America/New_York"},
	{Name: "Los Angeles", Timezone: "America/Los_Angeles"},
	{Name: "San Francisco", Timezone: "America/Los_Angeles"},
	{Name: "Chicago", Timezone: "America/Chicago"},
	{Name: "Shanghai", Timezone: "Asia/Shanghai"},
	{Name: "Beijing", Timezone: "Asia/Shanghai"},
	{Name: "Moscow", Timezone: "Europe/Moscow"},
	{Name: "Sydney", Timezone: "Australia/Sydney"},
	{Name: "Dubai", Timezone: "Asia/Dubai"},
	{Name: "Singapore", Timezone: "Asia/Singapore"},
	{Name: "Hong Kong", Timezone: "Asia/Hong_Kong"},
	{Name: "Seoul", Timezone: "Asia/Seoul"},
	{Name: "Mumbai", Timezone: "Asia/Kolkata"},
	{Name: "Toronto", Timezone: "America/Toronto"},
	{Name: "Vancouver", Timezone: "America/Vancouver"},
	{Name: "Sao Paulo", Timezone: "America/Sao_Paulo"},
	{Name: "Buenos Aires", Timezone: "America/Argentina/Buenos_Aires"},
	{Name: "Istanbul", Timezone: "Europe/Istanbul"},
	{Name: "Cairo", Timezone: "Africa/Cairo"},
	{Name: "Cape Town", Timezone: "Africa/Johannesburg"},
	{Name: "Nairobi", Timezone: "Africa/Nairobi"},
	{Name: "Zurich", Timezone: "Europe/Zurich"},
	{Name: "Amsterdam", Timezone: "Europe/Amsterdam"},
	{Name: "Madrid", Timezone: "Europe/Madrid"},
	{Name: "Rome", Timezone: "Europe/Rome"},
	{Name: "Stockholm", Timezone: "Europe/Stockholm"},
}

// GeoLocationByQuery looks up a city by name (case-insensitive).
// Returns the GeoLocation and true if found.
func GeoLocationByQuery(query string) (GeoLocation, bool) {
	for _, loc := range geoLocations {
		if strings.EqualFold(loc.Name, query) {
			return loc, true
		}
	}
	return GeoLocation{}, false
}
