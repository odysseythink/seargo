package engine

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
