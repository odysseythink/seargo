package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	answerer.Register(&weatherAnswerer{})
}

// KV is a simple key-value cache interface used by the weather answerer.
type KV interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

type weatherAnswerer struct{}

func (a *weatherAnswerer) Keywords() []string {
	return []string{"weather", "forecast", "wetter", "meteo", "météo", "tiempo", "tempo"}
}

func (a *weatherAnswerer) Info() answerer.AnswererInfo {
	return answerer.AnswererInfo{
		Name:        "weather",
		Description: "Current weather and forecast for a location",
		Keywords:    a.Keywords(),
		Examples: []string{
			"weather Paris",
			"forecast Tokyo",
		},
	}
}

func (a *weatherAnswerer) Answer(ctx *answerer.AnswerContext) []models.Result {
	locationQuery := stripLeadingKeyword(ctx.Query)
	if locationQuery == "" {
		return nil
	}

	c := getClient()
	if c == nil {
		return nil
	}
	kv := getCache()

	locale := resolveLocale(ctx.Locale)
	geo, ok := geocodeWithCache(c, kv, locationQuery)
	if !ok {
		return nil
	}

	weather, ok := fetchWeatherWithCache(c, kv, geo, locale)
	if !ok {
		return nil
	}

	symbol := symbolURL(c, kv, weather.Condition, weather.IsDay)
	infobox := buildWeatherInfobox(weather, symbol, locale)
	return results.ToAPIResult([]results.Result{&infobox})
}

var (
	clientMu   sync.RWMutex
	httpClient *httpx.Client
	cacheMu    sync.RWMutex
	cache      KV
)

// SetHTTPClient injects the HTTP client used by the weather answerer.
func SetHTTPClient(c *httpx.Client) {
	clientMu.Lock()
	defer clientMu.Unlock()
	httpClient = c
}

func getClient() *httpx.Client {
	clientMu.RLock()
	defer clientMu.RUnlock()
	return httpClient
}

// SetCache injects the key-value cache used by the weather answerer.
func SetCache(kv KV) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = kv
}

func getCache() KV {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return cache
}

var (
	openMeteoGeocodeURL  = "https://geocoding-api.open-meteo.com/v1/search"
	openMeteoForecastURL = "https://api.open-meteo.com/v1/forecast"
	wttrInBaseURL        = "https://wttr.in"
)

func stripLeadingKeyword(query string) string {
	query = strings.TrimSpace(query)
	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}
	first := strings.ToLower(words[0])
	for _, kw := range []string{"weather", "forecast", "wetter", "meteo", "météo", "tiempo", "tempo"} {
		if first == strings.ToLower(kw) {
			return strings.TrimSpace(strings.Join(words[1:], " "))
		}
	}
	return strings.TrimSpace(query)
}

func resolveLocale(locale string) string {
	if locale == "" {
		return "en"
	}
	return locale
}

func normalizeQuery(q string) string {
	return strings.ToLower(strings.TrimSpace(q))
}

func geocodeWithCache(c *httpx.Client, kv KV, query string) (GeoLocation, bool) {
	key := "geocode:" + normalizeQuery(query)
	if kv != nil {
		ctx := context.Background()
		if cached, ok, _ := kv.Get(ctx, key); ok && len(cached) > 0 {
			var geo GeoLocation
			if err := json.Unmarshal(cached, &geo); err == nil {
				return geo, true
			}
		}
	}

	geo, ok := geocodeOpenMeteo(c, query)
	if !ok {
		return GeoLocation{}, false
	}

	if kv != nil {
		ctx := context.Background()
		if data, err := json.Marshal(geo); err == nil {
			_ = kv.Set(ctx, key, data, 7*24*time.Hour)
		}
	}
	return geo, true
}

func geocodeOpenMeteo(c *httpx.Client, query string) (GeoLocation, bool) {
	u, err := url.Parse(openMeteoGeocodeURL)
	if err != nil {
		return GeoLocation{}, false
	}
	q := u.Query()
	q.Set("name", query)
	u.RawQuery = q.Encode()

	resp, err := c.R().SetTimeout(5 * time.Second).Get(u.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		return GeoLocation{}, false
	}

	var payload struct {
		Results []struct {
			Name        string  `json:"name"`
			Latitude    float64 `json:"latitude"`
			Longitude   float64 `json:"longitude"`
			Elevation   float64 `json:"elevation"`
			CountryCode string  `json:"country_code"`
			Timezone    string  `json:"timezone"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return GeoLocation{}, false
	}
	if len(payload.Results) == 0 {
		return GeoLocation{}, false
	}

	r := payload.Results[0]
	return GeoLocation{
		Name:        r.Name,
		Latitude:    r.Latitude,
		Longitude:   r.Longitude,
		Elevation:   r.Elevation,
		CountryCode: r.CountryCode,
		Timezone:    r.Timezone,
	}, true
}

func fetchWeatherWithCache(c *httpx.Client, kv KV, geo GeoLocation, locale string) (WeatherData, bool) {
	territory := territoryFromLocale(locale)
	key := fmt.Sprintf("weather:%.4f:%.4f:%s", geo.Latitude, geo.Longitude, territory)
	if kv != nil {
		ctx := context.Background()
		if cached, ok, _ := kv.Get(ctx, key); ok && len(cached) > 0 {
			var w WeatherData
			if err := json.Unmarshal(cached, &w); err == nil {
				return w, true
			}
		}
	}

	weather, ok := fetchOpenMeteo(c, geo)
	if !ok {
		weather, ok = fetchWttrIn(c, geo)
	}
	if !ok {
		return WeatherData{}, false
	}

	if kv != nil {
		ctx := context.Background()
		if data, err := json.Marshal(weather); err == nil {
			_ = kv.Set(ctx, key, data, 30*time.Minute)
		}
	}
	return weather, true
}

func fetchOpenMeteo(c *httpx.Client, geo GeoLocation) (WeatherData, bool) {
	u, err := url.Parse(openMeteoForecastURL)
	if err != nil {
		return WeatherData{}, false
	}
	q := u.Query()
	q.Set("latitude", strconv.FormatFloat(geo.Latitude, 'f', -1, 64))
	q.Set("longitude", strconv.FormatFloat(geo.Longitude, 'f', -1, 64))
	q.Set("current", "temperature_2m,apparent_temperature,is_day,weather_code,cloud_cover,relative_humidity_2m,wind_speed_10m,wind_direction_10m,surface_pressure")
	q.Set("forecast_days", "1")
	u.RawQuery = q.Encode()

	resp, err := c.R().SetTimeout(5 * time.Second).Get(u.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		return WeatherData{}, false
	}

	var payload struct {
		Current struct {
			Temperature2m       float64 `json:"temperature_2m"`
			ApparentTemperature float64 `json:"apparent_temperature"`
			IsDay               int     `json:"is_day"`
			WeatherCode         int     `json:"weather_code"`
			CloudCover          float64 `json:"cloud_cover"`
			RelativeHumidity2m  float64 `json:"relative_humidity_2m"`
			WindSpeed10m        float64 `json:"wind_speed_10m"`
			WindDirection10m    float64 `json:"wind_direction_10m"`
			SurfacePressure     float64 `json:"surface_pressure"`
		} `json:"current"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return WeatherData{}, false
	}

	return WeatherData{
		Location:      geo,
		Condition:     conditionFromWMO(payload.Current.WeatherCode),
		Temperature:   Temperature{Value: payload.Current.Temperature2m, Unit: "°C"},
		FeelsLike:     Temperature{Value: payload.Current.ApparentTemperature, Unit: "°C"},
		WindSpeed:     WindSpeed{Value: payload.Current.WindSpeed10m, Unit: "km/h"},
		WindDirection: Compass{Value: payload.Current.WindDirection10m, Unit: "°"},
		Pressure:      Pressure{Value: payload.Current.SurfacePressure, Unit: "hPa"},
		Humidity:      RelativeHumidity{Value: payload.Current.RelativeHumidity2m},
		CloudCover:    payload.Current.CloudCover,
		IsDay:         payload.Current.IsDay == 1,
		Forecasts:     nil,
	}, true
}

func fetchWttrIn(c *httpx.Client, geo GeoLocation) (WeatherData, bool) {
	u := fmt.Sprintf("%s/%s?format=j1", strings.TrimSuffix(wttrInBaseURL, "/"), url.PathEscape(geo.Name))
	resp, err := c.R().SetTimeout(5 * time.Second).Get(u)
	if err != nil || resp.StatusCode != http.StatusOK {
		return WeatherData{}, false
	}

	var payload struct {
		CurrentCondition []struct {
			TempC         string `json:"temp_C"`
			FeelsLikeC    string `json:"FeelsLikeC"`
			WindspeedKmph string `json:"windspeedKmph"`
			WinddirDegree string `json:"winddirDegree"`
			Pressure      string `json:"pressure"`
			Humidity      string `json:"humidity"`
			Cloudcover    string `json:"cloudcover"`
			WeatherCode   string `json:"weatherCode"`
			IsDay         string `json:"isday"`
			WeatherDesc   []struct {
				Value string `json:"value"`
			} `json:"weatherDesc"`
		} `json:"current_condition"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return WeatherData{}, false
	}
	if len(payload.CurrentCondition) == 0 {
		return WeatherData{}, false
	}

	cc := payload.CurrentCondition[0]
	condition := ""
	if len(cc.WeatherDesc) > 0 && cc.WeatherDesc[0].Value != "" {
		condition = strings.ToLower(cc.WeatherDesc[0].Value)
	} else {
		code, _ := strconv.Atoi(cc.WeatherCode)
		condition = conditionFromWMO(code)
	}

	return WeatherData{
		Location:      geo,
		Condition:     condition,
		Temperature:   Temperature{Value: parseFloat(cc.TempC), Unit: "°C"},
		FeelsLike:     Temperature{Value: parseFloat(cc.FeelsLikeC), Unit: "°C"},
		WindSpeed:     WindSpeed{Value: parseFloat(cc.WindspeedKmph), Unit: "km/h"},
		WindDirection: Compass{Value: parseFloat(cc.WinddirDegree), Unit: "°"},
		Pressure:      Pressure{Value: parseFloat(cc.Pressure), Unit: "hPa"},
		Humidity:      RelativeHumidity{Value: parseFloat(cc.Humidity)},
		CloudCover:    parseFloat(cc.Cloudcover),
		IsDay:         cc.IsDay != "0" && strings.ToLower(cc.IsDay) != "no",
		Forecasts:     nil,
	}, true
}

func buildWeatherInfobox(w WeatherData, symbol, locale string) results.InfoboxResult {
	ib := results.InfoboxResult{
		BaseResult: results.BaseResult{
			Title:    fmt.Sprintf("Weather in %s", w.Location.Name),
			Content:  w.Condition,
			Engine:   "weather",
			Template: "infobox",
		},
		InfoboxID: fmt.Sprintf("weather:%s", w.Location.Name),
		ImgSrc:    symbol,
		Attributes: []results.InfoboxAttribute{
			{Label: "Temperature", Value: w.Temperature.L10n(locale)},
			{Label: "Feels like", Value: w.FeelsLike.L10n(locale)},
			{Label: "Wind", Value: w.WindSpeed.L10n(locale)},
			{Label: "Pressure", Value: w.Pressure.L10n(locale)},
			{Label: "Humidity", Value: w.Humidity.L10n(locale)},
		},
		URLs: []results.InfoboxURL{
			{Title: "Open-Meteo", URL: "https://open-meteo.com"},
		},
	}
	return ib
}

func territoryFromLocale(locale string) string {
	parts := strings.Split(locale, "-")
	if len(parts) > 1 {
		return strings.ToUpper(parts[len(parts)-1])
	}
	return strings.ToUpper(locale)
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
