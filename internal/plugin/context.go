package plugin

// SearchContext carries the query and user state for plugin hook invocations.
type SearchContext struct {
	Query       string
	RawQuery    string
	Lang        string
	Locale      string
	SafeSearch  int
	PageNo      int
	TimeRange   string
	UserPlugins []string
	Preferences map[string]any
}

// AppContext carries application-level services available to plugins.
type AppContext struct {
	Config     interface{} // *config.Config — use interface to avoid import cycle
	Logger     interface{} // *logger.Logger
	HTTPClient interface{} // *httpx.Client
}
