package weather

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/seargo/seargo/internal/httpx"
)

// yrSymbolBaseURL is the base URL for NRK yr-weather-symbols SVG files.
// The upstream design referenced /dist/svg; the current repository layout uses
// /symbols/lightmode for the default light-mode icons.
var yrSymbolBaseURL = "https://raw.githubusercontent.com/nrkno/yr-weather-symbols/master/symbols/lightmode"

// symbolURL returns a base64 data URL for the SVG symbol matching the given
// condition. The result is cached in kv when available.
func symbolURL(c *httpx.Client, kv KV, condition string, isDay bool) string {
	filename := symbolFilename(condition, isDay)
	if filename == "" {
		return ""
	}

	if kv != nil {
		key := "weather_symbol:" + filename
		ctx := context.Background()
		if cached, ok, _ := kv.Get(ctx, key); ok && len(cached) > 0 {
			return string(cached)
		}
	}

	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(yrSymbolBaseURL, "/"), filename)
	resp, err := c.R().SetTimeout(5 * time.Second).Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}

	contentType := "image/svg+xml"
	if ct := resp.Headers["Content-Type"]; len(ct) > 0 && ct[0] != "" {
		contentType = ct[0]
	}
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(resp.Body))

	if kv != nil {
		key := "weather_symbol:" + filename
		ctx := context.Background()
		_ = kv.Set(ctx, key, []byte(dataURL), 7*24*time.Hour)
	}
	return dataURL
}

// symbolFilename maps a normalized condition string and day/night flag to an
// NRK yr-weather-symbols filename.
func symbolFilename(condition string, isDay bool) string {
	switch strings.ToLower(condition) {
	case "clear sky":
		if isDay {
			return "01d.svg"
		}
		return "01n.svg"
	case "mainly clear", "fair":
		if isDay {
			return "02d.svg"
		}
		return "02n.svg"
	case "partly cloudy":
		if isDay {
			return "03d.svg"
		}
		return "03n.svg"
	case "overcast", "cloudy":
		return "04.svg"
	case "fog", "depositing rime fog":
		return "15.svg"
	case "drizzle", "light drizzle", "moderate drizzle", "dense drizzle",
		"light freezing drizzle", "dense freezing drizzle":
		return "46.svg"
	case "rain", "light rain", "moderate rain", "heavy rain":
		return "09.svg"
	case "freezing rain", "light freezing rain", "heavy freezing rain",
		"sleet", "light sleet", "moderate sleet", "heavy sleet":
		return "12.svg"
	case "snow", "light snow", "moderate snow", "heavy snow", "snow grains",
		"light snow showers", "moderate snow showers", "heavy snow showers":
		return "13.svg"
	case "rain showers", "light rain showers", "moderate rain showers",
		"violent rain showers":
		if isDay {
			return "05d.svg"
		}
		return "05n.svg"
	case "thunderstorm", "thunderstorm with light hail", "thunderstorm with heavy hail",
		"light thunderstorm", "heavy thunderstorm":
		return "11.svg"
	default:
		if isDay {
			return "03d.svg"
		}
		return "03n.svg"
	}
}

// conditionFromWMO maps a WMO Weather interpretation code (WW) to a condition
// string. See https://open-meteo.com/en/docs.
func conditionFromWMO(code int) string {
	switch code {
	case 0:
		return "clear sky"
	case 1:
		return "mainly clear"
	case 2:
		return "partly cloudy"
	case 3:
		return "overcast"
	case 45:
		return "fog"
	case 48:
		return "depositing rime fog"
	case 51:
		return "light drizzle"
	case 53:
		return "moderate drizzle"
	case 55:
		return "dense drizzle"
	case 56:
		return "light freezing drizzle"
	case 57:
		return "dense freezing drizzle"
	case 61:
		return "light rain"
	case 63:
		return "moderate rain"
	case 65:
		return "heavy rain"
	case 66:
		return "light freezing rain"
	case 67:
		return "heavy freezing rain"
	case 71:
		return "light snow"
	case 73:
		return "moderate snow"
	case 75:
		return "heavy snow"
	case 77:
		return "snow grains"
	case 80:
		return "light rain showers"
	case 81:
		return "moderate rain showers"
	case 82:
		return "violent rain showers"
	case 85:
		return "light snow showers"
	case 86:
		return "heavy snow showers"
	case 95:
		return "thunderstorm"
	case 96:
		return "thunderstorm with light hail"
	case 99:
		return "thunderstorm with heavy hail"
	default:
		return "partly cloudy"
	}
}
