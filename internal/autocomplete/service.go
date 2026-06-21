package autocomplete

import (
	"context"
	"fmt"
	"strings"

	"github.com/seargo/seargo/internal/httpx"
)

const MinQueryLength = 2
const MaxQueryLength = 256

type Service struct {
	client *httpx.Client
	cache  *ResultCache
}

func NewService(client *httpx.Client, cache *ResultCache) *Service {
	if client == nil {
		panic("autocomplete.NewService: client must not be nil")
	}
	if cache == nil {
		cache = NewResultCache(DefaultCacheTTL)
	}
	return &Service{client: client, cache: cache}
}

func (s *Service) Suggest(ctx context.Context, backend string, query string, locale string) []string {
	query = normalizeQuery(query)
	if len(query) < MinQueryLength {
		return nil
	}

	provider, ok := Get(backend)
	if !ok {
		return nil
	}

	cacheKey := "ac:" + backend + ":" + locale + ":" + query
	if cached, hit := s.cache.Get(cacheKey); hit {
		return cached
	}

	var results []string
	func() {
		defer func() {
			if r := recover(); r != nil {
				results = nil
			}
		}()
		var err error
		results, err = provider.Fetch(ctx, query, locale)
		if err != nil {
			results = nil
		}
	}()

	if len(results) > 0 {
		s.cache.Set(cacheKey, results)
	}
	return results
}

func (s *Service) Cache() *ResultCache { return s.cache }

func normalizeQuery(q string) string {
	q = strings.TrimSpace(q)
	if len(q) > MaxQueryLength {
		q = q[:MaxQueryLength]
	}
	return q
}

// --- Locale helpers ---

func LocaleToLanguage(locale string) string {
	if idx := strings.IndexAny(locale, "_-"); idx >= 0 {
		return strings.ToLower(locale[:idx])
	}
	return strings.ToLower(locale)
}

func LocaleToCountry(locale string) string {
	if idx := strings.IndexAny(locale, "_-"); idx >= 0 && idx+1 < len(locale) {
		return strings.ToUpper(locale[idx+1:])
	}
	return ""
}

func LocaleToGoogleHL(locale string) string { return LocaleToLanguage(locale) }

func LocaleToGoogleSubdomain(locale string) string {
	lang := LocaleToLanguage(locale)
	subdomains := map[string]string{
		"en": "www.google.com", "de": "www.google.de", "fr": "www.google.fr",
		"ja": "www.google.co.jp", "zh": "www.google.com.hk", "ko": "www.google.co.kr",
		"ru": "www.google.ru", "es": "www.google.es", "pt": "www.google.pt",
		"it": "www.google.it", "nl": "www.google.nl", "pl": "www.google.pl",
		"tr": "www.google.com.tr", "ar": "www.google.com.sa",
	}
	if s, ok := subdomains[lang]; ok {
		return s
	}
	return "www.google.com"
}

func LocaleToDDGRegion(locale string) string {
	country := LocaleToCountry(locale)
	lang := LocaleToLanguage(locale)
	if country == "" {
		country = lang
	}
	return strings.ToLower(country) + "-" + lang
}

func LocaleToQwantLocale(locale string) string {
	lang := LocaleToLanguage(locale)
	country := LocaleToCountry(locale)
	if country == "" {
		country = strings.ToUpper(lang)
	}
	return lang + "_" + strings.ToUpper(country)
}

var startpageLangMap = map[string]string{
	"da": "dansk", "de": "deutsch", "en": "english",
	"es": "espanol", "fr": "francais", "nb": "norsk",
	"nl": "nederlands", "pl": "polski", "pt": "portugues",
	"sv": "svenska",
}

func LocaleToStartpageLanguage(locale string) string {
	lang := LocaleToLanguage(locale)
	if s, ok := startpageLangMap[lang]; ok {
		return s
	}
	return "english"
}

func LocaleToWikipediaLang(locale string) string {
	lang := LocaleToLanguage(locale)
	known := map[string]string{
		"zh": "zh", "en": "en", "de": "de", "fr": "fr",
		"ja": "ja", "ko": "ko", "ru": "ru", "es": "es",
		"pt": "pt", "it": "it", "nl": "nl", "pl": "pl",
		"sv": "sv", "ar": "ar", "tr": "tr", "vi": "vi",
		"th": "th", "id": "id", "cs": "cs", "fi": "fi",
		"hu": "hu", "ro": "ro", "uk": "uk", "he": "he",
		"da": "da", "no": "no", "sk": "sk", "bg": "bg",
		"ca": "ca", "el": "el", "fa": "fa", "hi": "hi",
		"hr": "hr", "lt": "lt", "ms": "ms", "sl": "sl",
		"sr": "sr", "et": "et", "lv": "lv", "tl": "tl",
	}
	if _, ok := known[lang]; ok {
		return lang
	}
	return "en"
}

func LocaleToWikipediaNetloc(locale string) string {
	lang := LocaleToWikipediaLang(locale)
	special := map[string]string{
		"commons":  "commons.wikimedia.org",
		"meta":     "meta.wikimedia.org",
		"species":  "species.wikimedia.org",
		"wikidata": "www.wikidata.org",
	}
	if s, ok := special[lang]; ok {
		return s
	}
	return fmt.Sprintf("%s.wikipedia.org", lang)
}
