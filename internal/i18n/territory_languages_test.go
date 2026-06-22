package i18n

import (
	"testing"
)

func TestTerritoryLanguages_Official(t *testing.T) {
	tl := DefaultTerritoryLanguages()

	frLangs := tl.Official["FR"]
	found := false
	for _, l := range frLangs {
		if l == "fr" {
			found = true
			break
		}
	}
	if !found {
		t.Error("FR should have French (fr) as an official language")
	}

	cnLangs := tl.Official["CN"]
	found = false
	for _, l := range cnLangs {
		if l == "zh" {
			found = true
			break
		}
	}
	if !found {
		t.Error("CN should have Chinese (zh) as an official language")
	}
}

func TestTerritoryLanguages_ByLanguage(t *testing.T) {
	tl := DefaultTerritoryLanguages()

	frInfos := tl.ByLanguage["fr"]
	if len(frInfos) == 0 {
		t.Fatal("fr should have territory entries")
	}

	foundFR := false
	for _, info := range frInfos {
		if info.Territory == "FR" {
			foundFR = true
			if info.PopulationPercent < 80 {
				t.Errorf("FR population_percent for fr = %.1f, expected >= 80", info.PopulationPercent)
			}
			break
		}
	}
	if !foundFR {
		t.Error("FR territory should be in fr's ByLanguage list")
	}

	for i := 1; i < len(frInfos); i++ {
		if frInfos[i].PopulationPercent > frInfos[i-1].PopulationPercent {
			t.Errorf("ByLanguage[fr] not sorted descending at index %d: %.1f > %.1f",
				i, frInfos[i].PopulationPercent, frInfos[i-1].PopulationPercent)
		}
	}
}

func TestTerritoryLanguages_ENToUS(t *testing.T) {
	tl := DefaultTerritoryLanguages()
	enInfos := tl.ByLanguage["en"]
	foundUS := false
	for _, info := range enInfos {
		if info.Territory == "US" {
			foundUS = true
			break
		}
	}
	if !foundUS {
		t.Error("ByLanguage[en] should include US territory")
	}
}
