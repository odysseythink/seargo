package engine

import (
	"strings"

	"github.com/seargo/seargo/internal/i18n"
)

// EngineTraits holds language and region mappings for an engine,
// ported from SearXNG's traits system.
type EngineTraits struct {
	DataType  string            `json:"data_type"`
	Languages map[string]string `json:"languages"`
	Regions   map[string]string `json:"regions"`
	AllLocale string            `json:"all_locale"`
	Custom    map[string]any    `json:"custom,omitempty"`
}

// EngineTraitsMap is a map of engine name to EngineTraits.
type EngineTraitsMap map[string]EngineTraits

// Lookup returns the traits for an engine, falling back to an empty
// EngineTraits if not found.
func (tm EngineTraitsMap) Lookup(name string) (EngineTraits, bool) {
	if tm == nil {
		return EngineTraits{}, false
	}
	t, ok := tm[name]
	if !ok {
		return EngineTraits{}, false
	}
	return t, true
}

// resolveTraits filters the traits based on configured language and region.
// If cfgLang is non-empty, only matching language entries are kept.
// If cfgRegion is non-empty, only matching region entries are kept.
func resolveTraits(traits EngineTraits, cfgLang, cfgRegion string) EngineTraits {
	result := EngineTraits{
		DataType:  traits.DataType,
		AllLocale: traits.AllLocale,
		Custom:    traits.Custom,
	}

	// Filter languages
	if traits.Languages != nil {
		if cfgLang != "" {
			result.Languages = make(map[string]string)
			if v, ok := traits.Languages[cfgLang]; ok {
				result.Languages[cfgLang] = v
			}
		} else {
			result.Languages = make(map[string]string, len(traits.Languages))
			for k, v := range traits.Languages {
				result.Languages[k] = v
			}
		}
	} else {
		result.Languages = make(map[string]string)
	}

	// Filter regions
	if traits.Regions != nil {
		if cfgRegion != "" {
			result.Regions = make(map[string]string)
			if v, ok := traits.Regions[cfgRegion]; ok {
				result.Regions[cfgRegion] = v
			}
		} else {
			result.Regions = make(map[string]string, len(traits.Regions))
			for k, v := range traits.Regions {
				result.Regions[k] = v
			}
		}
	} else {
		result.Regions = make(map[string]string)
	}

	return result
}

// ResolvedLocale is the result of resolving a user locale against an engine's
// language and region maps.
type ResolvedLocale struct {
	Language string // engine-specific language parameter value
	Region   string // engine-specific region parameter value
	All      bool   // use engine's "all" locale
}

// Resolve applies the full SearXNG get_engine_locale algorithm against this
// engine's language and region maps to find the best match for a user locale.
func (t EngineTraits) Resolve(userLocale string) ResolvedLocale {
	result := ResolvedLocale{}

	tl := i18n.DefaultTerritoryLanguages()

	// Resolve language
	if len(t.Languages) > 0 {
		langKeys := make([]string, 0, len(t.Languages))
		for k := range t.Languages {
			langKeys = append(langKeys, k)
		}
		langLocales := i18n.BuildEngineLocales(langKeys)
		matched := i18n.GetEngineLocale(userLocale, langLocales, "", tl)
		if matched != "" {
			if v, ok := t.Languages[matched]; ok {
				result.Language = v
			}
		}
		// Fallback: try language-only match (SearXNG step 6)
		if matched == "" {
			langCode := strings.SplitN(userLocale, "-", 2)[0]
			langCode = strings.SplitN(langCode, "_", 2)[0]
			if v, ok := t.Languages[langCode]; ok {
				result.Language = v
			}
		}
	}

	// Resolve region
	if len(t.Regions) > 0 {
		regionKeys := make([]string, 0, len(t.Regions))
		for k := range t.Regions {
			regionKeys = append(regionKeys, k)
		}
		regionLocales := i18n.BuildEngineLocales(regionKeys)
		matched := i18n.GetEngineLocale(userLocale, regionLocales, "", tl)
		if matched != "" {
			if v, ok := t.Regions[matched]; ok {
				result.Region = v
			}
		}
	}

	// All-locale fallback
	if result.Language == "" && result.Region == "" && t.AllLocale != "" && t.AllLocale != "null" {
		result.All = true
	}

	return result
}
