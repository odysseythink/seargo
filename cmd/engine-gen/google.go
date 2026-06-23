package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/seargo/seargo/internal/engine"
)

// skipCountries excludes Google region codes that are not useful as user regions.
var skipCountries = map[string]bool{
	"CAT": true, "EU": true, "UN": true,
}

func runGoogleFetch(outputPath string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	traits, err := fetchGoogleTraits(client)
	if err != nil {
		return fmt.Errorf("fetch google traits: %w", err)
	}
	return mergeGoogleTraits(outputPath, traits)
}

func fetchGoogleTraits(client *http.Client) (engine.EngineTraits, error) {
	languages, err := fetchGoogleLanguages(client)
	if err != nil {
		return engine.EngineTraits{}, fmt.Errorf("languages: %w", err)
	}
	regions, err := fetchGoogleRegions(client)
	if err != nil {
		return engine.EngineTraits{}, fmt.Errorf("regions: %w", err)
	}
	domains, err := fetchSupportedDomains(client)
	if err != nil {
		return engine.EngineTraits{}, fmt.Errorf("domains: %w", err)
	}

	return engine.EngineTraits{
		DataType:  "traits_v1",
		Languages: languages,
		Regions:   regions,
		AllLocale: "ZZ",
		Custom: map[string]any{
			"supported_domains": domains,
		},
	}, nil
}

func fetchGoogleLanguages(client *http.Client) (map[string]string, error) {
	resp, err := client.Get("https://www.google.com/preferences")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseGoogleLanguages(resp.Body)
}

func parseGoogleLanguages(r io.Reader) (map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	doc.Find("select[name='hl'] option").Each(func(_ int, s *goquery.Selection) {
		val, ok := s.Attr("value")
		if !ok || val == "" {
			return
		}
		out[mapLanguage(val)] = "lang_" + val
	})
	return out, nil
}

func fetchGoogleRegions(client *http.Client) (map[string]string, error) {
	resp, err := client.Get("https://www.google.com/preferences")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseGoogleRegions(resp.Body)
}

func parseGoogleRegions(r io.Reader) (map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	doc.Find("select[name='gl'] option").Each(func(_ int, s *goquery.Selection) {
		val, ok := s.Attr("value")
		if !ok || val == "" || skipCountries[val] {
			return
		}
		out[mapRegion(val)] = val
	})
	return out, nil
}

func fetchSupportedDomains(client *http.Client) (map[string]string, error) {
	resp, err := client.Get("https://www.google.com/supported_domains")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseSupportedDomains(resp.Body)
}

func parseSupportedDomains(r io.Reader) (map[string]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ".")
		region := strings.ToUpper(parts[len(parts)-1])
		out[region] = "www." + strings.TrimPrefix(line, ".")
	}
	return out, nil
}

func mergeGoogleTraits(outputPath string, traits engine.EngineTraits) error {
	existing := make(engine.EngineTraitsMap)
	if data, err := os.ReadFile(outputPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}
	existing["google"] = traits

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, append(out, '\n'), 0644)
}

// mapLanguage and mapRegion mirror SearXNG's normalization; identity by default.
func mapLanguage(v string) string { return v }
func mapRegion(v string) string   { return v }
