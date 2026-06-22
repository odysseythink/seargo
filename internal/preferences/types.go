package preferences

// PreferenceValue is a union-like representation of a single preference field.
type PreferenceValue struct {
	StringValue string
	StringSlice []string
	BoolValue   bool
	IntValue    int
}

// UserPreferences is the public read-only view returned by the API.
type UserPreferences struct {
	Categories             []string `json:"categories"`
	Language               string   `json:"language"`
	Locale                 string   `json:"locale"`
	Autocomplete           string   `json:"autocomplete"`
	FaviconResolver        string   `json:"favicon_resolver"`
	ImageProxy             bool     `json:"image_proxy"`
	Method                 string   `json:"method"`
	SafeSearch             int      `json:"safesearch"`
	Theme                  string   `json:"theme"`
	ResultsOnNewTab        bool     `json:"results_on_new_tab"`
	DOIResolver            string   `json:"doi_resolver"`
	SimpleStyle            string   `json:"simple_style"`
	CenterAlignment        bool     `json:"center_alignment"`
	QueryInTitle           bool     `json:"query_in_title"`
	SearchOnCategorySelect bool     `json:"search_on_category_select"`
	Hotkeys                string   `json:"hotkeys"`
	URLFormatting          string   `json:"url_formatting"`
	DisabledEngines        []string `json:"disabled_engines"`
	EnabledEngines         []string `json:"enabled_engines"`
	DisabledPlugins        []string `json:"disabled_plugins"`
	EnabledPlugins         []string `json:"enabled_plugins"`
	Tokens                 []string `json:"tokens"`
	Locked                 []string `json:"locked"`
}

// PreferencesUpdate is the request body for PUT /api/preferences.
// Pointer fields distinguish "not provided" (nil) from "set to empty".
type PreferencesUpdate struct {
	Categories             []string `json:"categories,omitempty"`
	Language               *string  `json:"language,omitempty"`
	Locale                 *string  `json:"locale,omitempty"`
	Autocomplete           *string  `json:"autocomplete,omitempty"`
	FaviconResolver        *string  `json:"favicon_resolver,omitempty"`
	ImageProxy             *bool    `json:"image_proxy,omitempty"`
	Method                 *string  `json:"method,omitempty"`
	SafeSearch             *int     `json:"safesearch,omitempty"`
	Theme                  *string  `json:"theme,omitempty"`
	ResultsOnNewTab        *bool    `json:"results_on_new_tab,omitempty"`
	DOIResolver            *string  `json:"doi_resolver,omitempty"`
	SimpleStyle            *string  `json:"simple_style,omitempty"`
	CenterAlignment        *bool    `json:"center_alignment,omitempty"`
	QueryInTitle           *bool    `json:"query_in_title,omitempty"`
	SearchOnCategorySelect *bool    `json:"search_on_category_select,omitempty"`
	Hotkeys                *string  `json:"hotkeys,omitempty"`
	URLFormatting          *string  `json:"url_formatting,omitempty"`
	DisabledEngines        []string `json:"disabled_engines,omitempty"`
	EnabledEngines         []string `json:"enabled_engines,omitempty"`
	DisabledPlugins        []string `json:"disabled_plugins,omitempty"`
	EnabledPlugins         []string `json:"enabled_plugins,omitempty"`
	Tokens                 []string `json:"tokens,omitempty"`
}
