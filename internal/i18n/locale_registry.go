package i18n

// LocaleInfo describes a single supported locale.
type LocaleInfo struct {
	Tag         string `json:"tag"`
	Name        string `json:"name"`
	RegionName  string `json:"region_name,omitempty"`
	EnglishName string `json:"english_name"`
	Flag        string `json:"flag"`
	RTL         bool   `json:"rtl"`
}

// LocaleRegistry holds the set of supported locales and RTL flags.
type LocaleRegistry struct {
	Supported  []LocaleInfo
	rtlLocales map[string]bool
	names      map[string]string
}

// NewLocaleRegistry creates a LocaleRegistry with SearXNG-derived locale data.
func NewLocaleRegistry() *LocaleRegistry {
	supported := []LocaleInfo{
		{Tag: "en", Name: "English", EnglishName: "English", Flag: "🇬🇧", RTL: false},
		{Tag: "zh-CN", Name: "简体中文", EnglishName: "Chinese (Simplified)", Flag: "🇨🇳", RTL: false},
		{Tag: "zh-TW", Name: "繁體中文", EnglishName: "Chinese (Traditional)", Flag: "🇹🇼", RTL: false},
		{Tag: "ar", Name: "العربية", EnglishName: "Arabic", Flag: "🇸🇦", RTL: true},
		{Tag: "he", Name: "עברית", EnglishName: "Hebrew", Flag: "🇮🇱", RTL: true},
	}

	r := &LocaleRegistry{
		Supported:  supported,
		rtlLocales: make(map[string]bool, len(supported)),
		names:      make(map[string]string, len(supported)),
	}

	for _, loc := range supported {
		r.rtlLocales[loc.Tag] = loc.RTL
		r.names[loc.Tag] = loc.Name
	}

	return r
}

// IsSupported checks whether the given locale tag is in the supported set.
func (r *LocaleRegistry) IsSupported(tag string) bool {
	_, ok := r.names[tag]
	return ok
}

// IsRTL checks whether the given locale tag uses right-to-left text direction.
func (r *LocaleRegistry) IsRTL(tag string) bool {
	return r.rtlLocales[tag]
}

// BestMatch selects the best supported locale from a list of Accept-Language tags.
// Returns the first match in tag order; falls back to "en".
func (r *LocaleRegistry) BestMatch(tags []string) string {
	for _, tag := range tags {
		if r.IsSupported(tag) {
			return tag
		}
	}
	return "en"
}
