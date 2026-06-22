package weather

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models/results"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// httpx logs outbound requests via mlog, which defaults to writing files
	// in the system temp directory. Route logs to stderr in tests to avoid
	// file creation conflicts and keep tests hermetic.
	_ = flag.Set("logtostderr", "true")
	os.Exit(m.Run())
}

// memoryKV is a simple in-memory cache implementation for tests.
type memoryKV struct {
	mu   sync.RWMutex
	data map[string][]byte
	ttl  map[string]time.Time
}

func newMemoryKV() *memoryKV {
	return &memoryKV{
		data: make(map[string][]byte),
		ttl:  make(map[string]time.Time),
	}
}

func (m *memoryKV) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if expiry, ok := m.ttl[key]; ok && time.Now().After(expiry) {
		return nil, false, nil
	}
	v, ok := m.data[key]
	return v, ok, nil
}

func (m *memoryKV) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	m.ttl[key] = time.Now().Add(ttl)
	return nil
}

func (m *memoryKV) count(keyPrefix string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for k := range m.data {
		if strings.HasPrefix(k, keyPrefix) {
			n++
		}
	}
	return n
}

// serverHits tracks how many times each mock endpoint was invoked.
type serverHits struct {
	mu       sync.Mutex
	geocode  int
	forecast int
	wttr     int
	symbol   int
}

func (h *serverHits) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.geocode = 0
	h.forecast = 0
	h.wttr = 0
	h.symbol = 0
}

func setupTest(t *testing.T) (*httptest.Server, *httpx.Client, *memoryKV, *answerer.AnswererStorage, *serverHits) {
	t.Helper()

	hits := &serverHits{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasPrefix(path, "/geocode"):
			hits.mu.Lock()
			hits.geocode++
			hits.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{
						"name":         "Paris",
						"latitude":     48.8567,
						"longitude":    2.3510,
						"elevation":    35.0,
						"country_code": "FR",
						"timezone":     "Europe/Paris",
					},
				},
			})
		case strings.HasPrefix(path, "/forecast"):
			hits.mu.Lock()
			hits.forecast++
			hits.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"current": map[string]any{
					"temperature_2m":       22.5,
					"apparent_temperature": 21.0,
					"is_day":               1,
					"weather_code":         2,
					"cloud_cover":          30.0,
					"relative_humidity_2m": 55.0,
					"wind_speed_10m":       12.5,
					"wind_direction_10m":   270.0,
					"surface_pressure":     1015.0,
				},
			})
		case strings.HasPrefix(path, "/wttr/"):
			hits.mu.Lock()
			hits.wttr++
			hits.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"current_condition": []map[string]any{
					{
						"temp_C":        "18",
						"FeelsLikeC":    "17",
						"windspeedKmph": "10",
						"winddirDegree": "180",
						"pressure":      "1012",
						"humidity":      "60",
						"cloudcover":    "40",
						"weatherCode":   "61",
						"isday":         "1",
						"weatherDesc": []map[string]string{
							{"value": "light rain"},
						},
					},
				},
			})
		case strings.HasPrefix(path, "/symbol/"):
			hits.mu.Lock()
			hits.symbol++
			hits.mu.Unlock()
			w.Header().Set("Content-Type", "image/svg+xml")
			_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle/></svg>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(ts.Close)

	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			EnableHTTP: true,
		},
	}
	registry, err := httpx.NewRegistry(cfg)
	require.NoError(t, err)
	client := httpx.NewClient(registry, "default", "weather", "seargo-weather/1.0", 5*time.Second)

	kv := newMemoryKV()
	storage := answerer.NewAnswererStorage()

	// Override service URLs to point to the test server.
	openMeteoGeocodeURL = ts.URL + "/geocode"
	openMeteoForecastURL = ts.URL + "/forecast"
	wttrInBaseURL = ts.URL + "/wttr"
	yrSymbolBaseURL = ts.URL + "/symbol"

	SetHTTPClient(client)
	SetCache(kv)

	return ts, client, kv, storage, hits
}

func TestWeatherAnswerer_OpenMeteo(t *testing.T) {
	_, _, kv, storage, hits := setupTest(t)

	a := &weatherAnswerer{}
	storage.Register(a)

	res := storage.Ask(&answerer.AnswerContext{Query: "weather Paris", Locale: "en"})
	require.Len(t, res, 1)
	assert.Equal(t, "infobox", res[0].Kind)
	assert.Equal(t, "Weather in Paris", res[0].Title)
	assert.Equal(t, "weather", res[0].Engine)

	extra, ok := res[0].Extra["attributes"].([]results.InfoboxAttribute)
	require.True(t, ok)
	require.Len(t, extra, 5)
	assert.Equal(t, "Temperature", extra[0].Label)
	assert.Equal(t, "22.5 °C", extra[0].Value)
	assert.Equal(t, "Feels like", extra[1].Label)
	assert.Equal(t, "21.0 °C", extra[1].Value)
	assert.Equal(t, "Wind", extra[2].Label)
	assert.Equal(t, "12.5 km/h", extra[2].Value)
	assert.Equal(t, "Pressure", extra[3].Label)
	assert.Equal(t, "1015.0 hPa", extra[3].Value)
	assert.Equal(t, "Humidity", extra[4].Label)
	assert.Equal(t, "55%", extra[4].Value)

	assert.Equal(t, 1, hits.geocode)
	assert.Equal(t, 1, hits.forecast)
	assert.Equal(t, 1, hits.symbol)
	assert.Equal(t, 0, hits.wttr)

	assert.Equal(t, 1, kv.count("geocode:"))
	assert.Equal(t, 1, kv.count("weather:"))
	assert.Equal(t, 1, kv.count("weather_symbol:"))
}

func TestWeatherAnswerer_USLocale(t *testing.T) {
	_, _, _, storage, _ := setupTest(t)

	a := &weatherAnswerer{}
	storage.Register(a)

	res := storage.Ask(&answerer.AnswerContext{Query: "weather Paris", Locale: "en-US"})
	require.Len(t, res, 1)

	extra, ok := res[0].Extra["attributes"].([]results.InfoboxAttribute)
	require.True(t, ok)
	assert.Equal(t, "72.5 °F", extra[0].Value)
}

func TestWeatherAnswerer_CacheHit(t *testing.T) {
	_, _, _, storage, hits := setupTest(t)

	a := &weatherAnswerer{}
	storage.Register(a)

	// First request populates cache.
	res1 := storage.Ask(&answerer.AnswerContext{Query: "weather Paris"})
	require.Len(t, res1, 1)
	assert.Equal(t, 1, hits.geocode)
	assert.Equal(t, 1, hits.forecast)
	assert.Equal(t, 1, hits.symbol)

	hits.reset()

	// Second request should hit cache.
	res2 := storage.Ask(&answerer.AnswerContext{Query: "weather Paris"})
	require.Len(t, res2, 1)
	assert.Equal(t, res1[0].Title, res2[0].Title)
	assert.Equal(t, 0, hits.geocode)
	assert.Equal(t, 0, hits.forecast)
	assert.Equal(t, 0, hits.symbol)
}

func TestWeatherAnswerer_FallbackWttrIn(t *testing.T) {
	_, _, _, storage, hits := setupTest(t)

	// Force open-meteo forecast to fail.
	openMeteoForecastURL = "http://127.0.0.1:1/forecast"

	a := &weatherAnswerer{}
	storage.Register(a)

	res := storage.Ask(&answerer.AnswerContext{Query: "weather Paris"})
	require.Len(t, res, 1)
	assert.Equal(t, "Weather in Paris", res[0].Title)
	assert.Equal(t, 0, hits.forecast)
	assert.Equal(t, 1, hits.wttr)

	extra, ok := res[0].Extra["attributes"].([]results.InfoboxAttribute)
	require.True(t, ok)
	assert.Equal(t, "18.0 °C", extra[0].Value)
}

func TestWeatherAnswerer_NoResultsWhenAllFail(t *testing.T) {
	_, _, _, storage, _ := setupTest(t)

	// Force both weather endpoints to fail.
	openMeteoForecastURL = "http://127.0.0.1:1/forecast"
	wttrInBaseURL = "http://127.0.0.1:1/wttr"

	a := &weatherAnswerer{}
	storage.Register(a)

	res := storage.Ask(&answerer.AnswerContext{Query: "weather Paris"})
	assert.Empty(t, res)
}

func TestWeatherAnswerer_KeywordStripping(t *testing.T) {
	_, _, _, storage, _ := setupTest(t)

	a := &weatherAnswerer{}
	storage.Register(a)

	tests := []struct {
		query    string
		expected string
	}{
		{"weather Paris", "Paris"},
		{"forecast Paris", "Paris"},
		{"WETTER Berlin", "Berlin"},
		{"météo Lyon", "Lyon"},
		{"tiempo Madrid", "Madrid"},
		{"tempo Roma", "Roma"},
		{"weather in Paris", "in Paris"},
	}

	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			got := stripLeadingKeyword(tc.query)
			assert.Equal(t, tc.expected, got)
		})
	}

	// Only keyword, no location.
	res := storage.Ask(&answerer.AnswerContext{Query: "weather"})
	assert.Empty(t, res)
}

func TestWeatherAnswerer_Keywords(t *testing.T) {
	a := &weatherAnswerer{}
	assert.ElementsMatch(t, []string{"weather", "forecast", "wetter", "meteo", "météo", "tiempo", "tempo"}, a.Keywords())
}

func TestWeatherAnswerer_Info(t *testing.T) {
	a := &weatherAnswerer{}
	info := a.Info()
	assert.Equal(t, "weather", info.Name)
	assert.NotEmpty(t, info.Description)
}

func TestWeatherAnswerer_NoClient(t *testing.T) {
	oldClient := getClient()
	SetHTTPClient(nil)
	defer SetHTTPClient(oldClient)

	a := &weatherAnswerer{}
	res := a.Answer(&answerer.AnswerContext{Query: "weather Paris"})
	assert.Nil(t, res)
}

func TestStripLeadingKeyword_ParisWeather(t *testing.T) {
	// The answerer only strips leading keywords, so "paris weather" stays as-is.
	got := stripLeadingKeyword("paris weather")
	assert.Equal(t, "paris weather", got)
}

func TestConditionFromWMO(t *testing.T) {
	assert.Equal(t, "clear sky", conditionFromWMO(0))
	assert.Equal(t, "mainly clear", conditionFromWMO(1))
	assert.Equal(t, "partly cloudy", conditionFromWMO(2))
	assert.Equal(t, "overcast", conditionFromWMO(3))
	assert.Equal(t, "fog", conditionFromWMO(45))
	assert.Equal(t, "moderate rain", conditionFromWMO(63))
	assert.Equal(t, "moderate snow", conditionFromWMO(73))
	assert.Equal(t, "thunderstorm", conditionFromWMO(95))
}

func TestSymbolFilename(t *testing.T) {
	assert.Equal(t, "01d.svg", symbolFilename("clear sky", true))
	assert.Equal(t, "01n.svg", symbolFilename("clear sky", false))
	assert.Equal(t, "04.svg", symbolFilename("overcast", true))
	assert.Equal(t, "15.svg", symbolFilename("fog", true))
	assert.Equal(t, "09.svg", symbolFilename("rain", true))
	assert.Equal(t, "13.svg", symbolFilename("snow", true))
	assert.Equal(t, "11.svg", symbolFilename("thunderstorm", true))
}

func TestTemperatureL10n(t *testing.T) {
	temp := Temperature{Value: 0, Unit: "°C"}
	assert.Equal(t, "32.0 °F", temp.L10n("en-US"))
	assert.Equal(t, "0.0 °C", temp.L10n("en"))
}
