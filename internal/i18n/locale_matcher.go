package i18n

import (
	"strings"

	"golang.org/x/text/language"
)

// explicitComponents determines whether a normalized locale tag explicitly
// contains a script and/or region subtag. We inspect the raw tag string
// because language.Parse adds inferred values and we cannot distinguish
// explicit from inferred via the exported API.
func explicitComponents(tag string) (hasScript, hasRegion bool) {
	parts := strings.Split(tag, "-")
	if len(parts) >= 3 {
		return true, true
	}
	if len(parts) == 2 {
		// Script subtags are always 4 letters; region are 2 letters or 3 digits.
		if len(parts[1]) == 4 {
			return true, false
		}
		return false, true
	}
	return false, false
}

// BuildEngineLocales builds a map from normalized SearGo locale tags to engine
// locale keys, implementing the SearXNG build_engine_locales logic.
func BuildEngineLocales(tagList []string) map[string]string {
	locales := make(map[string]string, len(tagList)*2)

	for _, tag := range tagList {
		tag = NormalizeLocaleTag(tag)
		locale, err := language.Parse(tag)
		if err != nil {
			continue
		}

		base, _ := locale.Base()
		baseStr := base.String()

		hasScript, hasRegion := explicitComponents(tag)

		if hasRegion || hasScript {
			if hasRegion {
				// Region is either parts[1] (language-region) or parts[2] (language-script-region)
				parts := strings.Split(tag, "-")
				regionStr := parts[len(parts)-1]
				regionTag := baseStr + "-" + regionStr
				locales[regionTag] = tag
			}
			if hasScript {
				// Script is always parts[1] in a normalized tag
				parts := strings.Split(tag, "-")
				scriptStr := parts[1]
				scriptTag := baseStr + "_" + scriptStr
				locales[scriptTag] = tag
			}
		} else {
			locales[baseStr] = tag
		}
	}

	return locales
}

// GetEngineLocale implements SearXNG's get_engine_locale: resolves a SearGo
// locale tag to the best matching engine-specific locale key.
func GetEngineLocale(userLocale string, engineLocales map[string]string, defaultLocale string, tl TerritoryLanguages) string {
	if len(engineLocales) == 0 {
		return defaultLocale
	}

	// 1. Direct 1:1 match
	normalized := NormalizeLocaleTag(userLocale)
	if v, ok := engineLocales[normalized]; ok {
		return v
	}

	// 2. Parse user locale
	loc, err := language.Parse(normalized)
	if err != nil {
		return defaultLocale
	}

	base, _ := loc.Base()
	script, _ := loc.Script()
	region, _ := loc.Region()

	// 3. Language-script mapping
	if script.String() != "" && script.String() != "Zzzz" {
		langScript := base.String() + "_" + script.String()
		if v, ok := engineLocales[langScript]; ok {
			return v
		}
	}

	// 4. Territory official languages fallback
	if region.String() != "" {
		regionStr := region.String()
		if officialLangs, ok := tl.Official[regionStr]; ok {
			for _, lang := range officialLangs {
				candidate := lang + "-" + regionStr
				if v, ok := engineLocales[candidate]; ok {
					return v
				}
			}
		}
	}

	// 5. Language official in other territories (population-weighted)
	if base.String() != "" {
		langStr := base.String()
		if infos, ok := tl.ByLanguage[langStr]; ok {
			// 5a. Prefer the "language as uppercase/ISO" territory
			// with the en→US special case
			preferred := strings.ToUpper(langStr)
			if preferred == "EN" {
				preferred = "US"
			}
			for _, info := range infos {
				if info.Territory == preferred {
					candidate := langStr + "-" + info.Territory
					if v, ok := engineLocales[candidate]; ok {
						return v
					}
				}
			}

			// 5b. Fall through population-sorted list
			for _, info := range infos {
				candidate := langStr + "-" + info.Territory
				if v, ok := engineLocales[candidate]; ok {
					return v
				}
			}
		}
	}

	return defaultLocale
}
