package externalurls

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExternalURL(t *testing.T) {
	// Ensure default catalog is loaded.
	require.NoError(t, Load("nonexistent_file_for_default_catalog.json"))

	tests := []struct {
		name        string
		urlID       string
		itemID      string
		alternative string
		want        string
	}{
		{
			name: "imdb_title default",
			urlID: "imdb_title", itemID: "tt123", alternative: "default",
			want: "https://www.imdb.com/title/tt123",
		},
		{
			name: "imdb_title mobile alternative",
			urlID: "imdb_title", itemID: "tt123", alternative: "mobile",
			want: "https://m.imdb.com/title/tt123",
		},
		{
			name: "unknown urlID",
			urlID: "unknown", itemID: "x", alternative: "default",
			want: "",
		},
		{
			name: "empty itemID returns template",
			urlID: "map", itemID: "", alternative: "default",
			want: "https://www.openstreetmap.org/?mlat=${latitude}&mlon=${longitude}&zoom=${zoom}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExternalURL(tt.urlID, tt.itemID, tt.alternative)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetIMDBURLID(t *testing.T) {
	tests := []struct {
		itemID string
		want   string
	}{
		{"tt123", "imdb_title"},
		{"mn456", "imdb_name"},
		{"ch789", "imdb_character"},
		{"co012", "imdb_company"},
		{"ev345", "imdb_event"},
		{"xx999", ""},
		{"x", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.itemID, func(t *testing.T) {
			assert.Equal(t, tt.want, GetIMDBURLID(tt.itemID))
		})
	}
}

func TestGetWikimediaImageID(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{
			raw:  "http://commons.wikimedia.org/wiki/Special:FilePath/Example.jpg",
			want: "Example.jpg",
		},
		{
			raw:  "File:Example.jpg",
			want: "Example.jpg",
		},
		{
			raw:  "Example.jpg",
			want: "Example.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			assert.Equal(t, tt.want, GetWikimediaImageID(tt.raw))
		})
	}
}

func TestGetEarthCoordinatesURL(t *testing.T) {
	url := GetEarthCoordinatesURL(52.5, 13.4, 12)
	assert.Contains(t, url, "mlat=52.5")
	assert.Contains(t, url, "mlon=13.4")
	assert.Contains(t, url, "zoom=12")
}

func TestAreaToOSMZoom(t *testing.T) {
	tests := []struct {
		areaKm2 float64
		want    int
	}{
		{-1, 19},
		{0, 19},
		{1e-6, 19},
		{1, 15},
		{100, 12},
		{1e15, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := AreaToOSMZoom(tt.areaKm2)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "external_urls.json")

	custom := `{
		"imdb_title": {
			"category_name": "IMDB",
			"url_name": "IMDB title",
			"urls": {
				"default": "https://example.com/title/$1"
			}
		}
	}`
	require.NoError(t, os.WriteFile(path, []byte(custom), 0644))

	require.NoError(t, Load(path))
	assert.Equal(t, "https://example.com/title/tt999", GetExternalURL("imdb_title", "tt999", "default"))
	assert.Equal(t, "", GetExternalURL("map", "", "default"))

	// Restore default catalog for any later tests.
	t.Cleanup(func() { _ = Load("nonexistent_file_for_default_catalog.json") })
}

func TestLoadMissingFileFallsBack(t *testing.T) {
	require.NoError(t, Load(filepath.Join(t.TempDir(), "missing.json")))
	assert.Equal(t, "https://www.imdb.com/title/tt123", GetExternalURL("imdb_title", "tt123", "default"))
}

func TestLoadRealCatalogFile(t *testing.T) {
	require.NoError(t, Load("../../../data/external_urls.json"))
	assert.Equal(t, "https://www.imdb.com/title/tt123", GetExternalURL("imdb_title", "tt123", "default"))
	assert.Equal(t, "https://x.com/seargo", GetExternalURL("twitter_profile", "seargo", "default"))
}

func TestGetWikimediaThumbnailURL(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "svg flag",
			raw:  "https://commons.wikimedia.org/wiki/Special:FilePath/Flag_of_Germany.svg?width=300",
			want: "https://upload.wikimedia.org/wikipedia/commons/thumb/b/ba/Flag_of_Germany.svg/300px-Flag_of_Germany.svg.png",
		},
		{
			name: "jpg photo",
			raw:  "https://commons.wikimedia.org/wiki/Special:FilePath/Albert_Einstein_Head.jpg?width=300",
			want: "https://upload.wikimedia.org/wikipedia/commons/thumb/d/d3/Albert_Einstein_Head.jpg/300px-Albert_Einstein_Head.jpg",
		},
		{
			name: "non-commons url is preserved",
			raw:  "https://example.com/image.png?width=300",
			want: "https://example.com/image.png?width=300",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, GetWikimediaThumbnailURL(tc.raw))
		})
	}
}
