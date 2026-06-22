package i18n

import (
	"testing"
)

func TestBuildEngineLocales(t *testing.T) {
	tagList := []string{"en", "en-US", "zh", "zh-CN", "zh-TW"}
	locales := BuildEngineLocales(tagList)

	// en → en
	if locales["en"] != "en" {
		t.Errorf("locales[en] = %q, want en", locales["en"])
	}
	// en-US → en-US
	if locales["en-US"] != "en-US" {
		t.Errorf("locales[en-US] = %q, want en-US", locales["en-US"])
	}
	// zh-CN should be built
	if _, ok := locales["zh-CN"]; !ok {
		t.Error("locales should include zh-CN")
	}
}

func TestGetEngineLocale_DirectMatch(t *testing.T) {
	tl := DefaultTerritoryLanguages()
	engineLocales := BuildEngineLocales([]string{"fr-BE", "fr-FR", "fr"})

	got := GetEngineLocale("fr-BE", engineLocales, "en", tl)
	if got != "fr-BE" {
		t.Errorf("direct match: got %q, want fr-BE", got)
	}
}

func TestGetEngineLocale_OfficialLanguageFallback(t *testing.T) {
	tl := DefaultTerritoryLanguages()
	engineLocales := BuildEngineLocales([]string{"fr-FR"})

	got := GetEngineLocale("fr-BE", engineLocales, "en", tl)
	if got != "fr-FR" {
		t.Errorf("official language fallback: got %q, want fr-FR", got)
	}
}

func TestGetEngineLocale_DefaultFallback(t *testing.T) {
	tl := DefaultTerritoryLanguages()
	engineLocales := BuildEngineLocales([]string{"de"})

	got := GetEngineLocale("xx-YY", engineLocales, "en", tl)
	if got != "en" {
		t.Errorf("unparseable locale: got %q, want en", got)
	}
}

func TestGetEngineLocale_PopulationWeightedFallback(t *testing.T) {
	tl := DefaultTerritoryLanguages()
	engineLocales := BuildEngineLocales([]string{"en-US"})

	got := GetEngineLocale("en", engineLocales, "en-GB", tl)
	if got != "en-US" {
		t.Errorf("population-weighted: got %q, want en-US", got)
	}
}

func TestGetEngineLocale_EmptyEngineLocales(t *testing.T) {
	tl := DefaultTerritoryLanguages()
	engineLocales := map[string]string{}

	got := GetEngineLocale("zh-CN", engineLocales, "en", tl)
	if got != "en" {
		t.Errorf("empty engine locales: got %q, want en (default)", got)
	}
}

func TestGetEngineLocale_ArWithoutDialect(t *testing.T) {
	tl := DefaultTerritoryLanguages()
	engineLocales := BuildEngineLocales([]string{"ar-EG"})

	got := GetEngineLocale("ar", engineLocales, "en", tl)
	if got != "ar-EG" {
		t.Errorf("ar population fallback: got %q, want ar-EG", got)
	}
}
