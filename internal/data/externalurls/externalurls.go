// Package externalurls provides a catalog of external URL templates used by
// infoboxes and the Wikidata engine.
package externalurls

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"
)

// Entry describes one external URL template group.
type Entry struct {
	CategoryName string            `json:"category_name"`
	URLName      string            `json:"url_name"`
	URLs         map[string]string `json:"urls"`
}

var catalog = builtInCatalog()

// builtInCatalog returns a minimal catalog used when external_urls.json is missing.
func builtInCatalog() map[string]Entry {
	return map[string]Entry{
		"imdb_title": {
			CategoryName: "IMDB",
			URLName:      "IMDB title",
			URLs: map[string]string{
				"default": "https://www.imdb.com/title/$1",
				"mobile":  "https://m.imdb.com/title/$1",
			},
		},
		"imdb_name": {
			CategoryName: "IMDB",
			URLName:      "IMDB name",
			URLs: map[string]string{
				"default": "https://www.imdb.com/name/$1",
				"mobile":  "https://m.imdb.com/name/$1",
			},
		},
		"map": {
			CategoryName: "OpenStreetMap",
			URLName:      "OpenStreetMap",
			URLs: map[string]string{
				"default": "https://www.openstreetmap.org/?mlat=${latitude}&mlon=${longitude}&zoom=${zoom}",
			},
		},
		"wikimedia_image": {
			CategoryName: "Wikimedia Commons",
			URLName:      "Wikimedia Commons image",
			URLs: map[string]string{
				"default": "https://commons.wikimedia.org/wiki/File:$1",
			},
		},
		"twitter_profile": {
			CategoryName: "X",
			URLName:      "X profile",
			URLs: map[string]string{
				"default": "https://x.com/$1",
			},
		},
		"facebook_profile": {
			CategoryName: "Facebook",
			URLName:      "Facebook profile",
			URLs: map[string]string{
				"default": "https://www.facebook.com/$1",
			},
		},
	}
}

// Load reads the external URL catalog from path. If the file is missing, the
// built-in minimal catalog is used.
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			catalog = builtInCatalog()
			return nil
		}
		return fmt.Errorf("read external urls: %w", err)
	}

	var loaded map[string]Entry
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("parse external urls: %w", err)
	}

	catalog = loaded
	return nil
}

// GetExternalURL returns the URL template for urlID filled with itemID.
// If alternative is empty or not found, the "default" alternative is used.
// An empty string is returned when the urlID is unknown or has no template.
func GetExternalURL(urlID, itemID, alternative string) string {
	entry, ok := catalog[urlID]
	if !ok {
		return ""
	}

	tmpl := entry.URLs[alternative]
	if tmpl == "" {
		tmpl = entry.URLs["default"]
	}
	if tmpl == "" {
		return ""
	}

	if itemID == "" {
		return tmpl
	}

	return strings.Replace(tmpl, "$1", itemID, -1)
}

// GetIMDBURLID maps an IMDB item ID prefix to its external URL ID.
func GetIMDBURLID(itemID string) string {
	if len(itemID) < 2 {
		return ""
	}

	switch itemID[:2] {
	case "tt":
		return "imdb_title"
	case "mn":
		return "imdb_name"
	case "ch":
		return "imdb_character"
	case "co":
		return "imdb_company"
	case "ev":
		return "imdb_event"
	}

	return ""
}

// GetWikimediaImageID strips Wikimedia Commons file path prefixes from raw.
func GetWikimediaImageID(raw string) string {
	prefix := "http://commons.wikimedia.org/wiki/Special:FilePath/"
	if strings.HasPrefix(raw, prefix) {
		return raw[len(prefix):]
	}
	if strings.HasPrefix(raw, "File:") {
		return raw[len("File:"):]
	}
	return raw
}

// GetEarthCoordinatesURL returns an OpenStreetMap URL for the given coordinates.
func GetEarthCoordinatesURL(lat, lon float64, zoom int) string {
	url := GetExternalURL("map", "", "default")
	if url == "" {
		return ""
	}

	url = strings.Replace(url, "${latitude}", fmt.Sprintf("%g", lat), -1)
	url = strings.Replace(url, "${longitude}", fmt.Sprintf("%g", lon), -1)
	url = strings.Replace(url, "${zoom}", fmt.Sprintf("%d", zoom), -1)
	return url
}

const wikimediaImageDefaultPrefix = "https://commons.wikimedia.org/wiki/Special:FilePath/"

// GetWikimediaThumbnailURL converts a Wikimedia Commons "Special:FilePath"
// URL into a static upload.wikimedia.org thumbnail URL. It follows the same
// MD5-path algorithm used by upstream SearXNG.
func GetWikimediaThumbnailURL(raw string) string {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return raw
	}
	first := fields[0]
	if !strings.Contains(first, wikimediaImageDefaultPrefix) {
		return raw
	}

	rest := strings.TrimPrefix(first, wikimediaImageDefaultPrefix)
	parts := strings.SplitN(rest, "?", 2)
	name := strings.ReplaceAll(parts[0], "%20", "_")
	name, _ = url.PathUnescape(name)
	if name == "" {
		return raw
	}

	nameFirst := name
	nameSecond := name
	if strings.Contains(strings.Fields(name)[0], ".svg") {
		nameSecond = name + ".png"
	}

	if len(parts) < 2 {
		return raw
	}
	sizePart := parts[1]
	eq := strings.Index(sizePart, "=")
	if eq < 0 {
		return raw
	}
	size := sizePart[eq+1:]
	if amp := strings.Index(size, "&"); amp >= 0 {
		size = size[:amp]
	}
	if size == "" {
		return raw
	}

	sum := fmt.Sprintf("%x", md5.Sum([]byte(name)))
	return fmt.Sprintf(
		"https://upload.wikimedia.org/wikipedia/commons/thumb/%s/%s/%s/%spx-%s",
		sum[:1], sum[:2], nameFirst, size, nameSecond,
	)
}

// AreaToOSMZoom converts an area in square kilometers to an OpenStreetMap zoom level.
func AreaToOSMZoom(areaKm2 float64) int {
	if areaKm2 <= 0 {
		return 19
	}

	zoom := int(math.Round(19 - 0.688297*math.Log(226.878*areaKm2)))
	if zoom < 0 {
		return 0
	}
	if zoom > 19 {
		return 19
	}
	return zoom
}
