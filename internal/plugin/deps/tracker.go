package deps

import (
	"net/url"
	"sync"
)

var (
	trackerPatterns []string
	trackerOnce     sync.Once
)

// InitTrackerPatterns initializes the list of known tracking parameters.
// It is safe for concurrent use and idempotent via sync.Once.
func InitTrackerPatterns() {
	trackerOnce.Do(func() {
		trackerPatterns = []string{
			"utm_source",
			"utm_medium",
			"utm_campaign",
			"utm_term",
			"utm_content",
			"fbclid",
			"gclid",
			"dclid",
			"msclkid",
			"twclid",
			"ref",
			"ref_src",
			"ref_url",
			"mc_cid",
			"mc_eid",
			"_ga",
			"_gl",
			"yclid",
			"_openstat",
		}
	})
}

// TrackerCleanURL removes known tracking query parameters from a URL.
// Returns the cleaned URL string and a boolean indicating whether any
// parameters were removed.
func TrackerCleanURL(rawURL string) (string, bool) {
	if rawURL == "" {
		return rawURL, false
	}

	InitTrackerPatterns()

	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, false
	}

	q := u.Query()
	removed := false
	for _, param := range trackerPatterns {
		if q.Get(param) != "" {
			q.Del(param)
			removed = true
		}
	}

	if !removed {
		return rawURL, false
	}

	u.RawQuery = q.Encode()
	return u.String(), true
}
