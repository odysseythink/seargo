package i18n

import (
	"sort"
	"strconv"
	"strings"
)

// NormalizeLocaleTag normalizes a locale tag to BCP 47 format:
// language lowercase, region uppercase, underscore→hyphen.
func NormalizeLocaleTag(tag string) string {
	tag = strings.ReplaceAll(tag, "_", "-")
	parts := strings.SplitN(tag, "-", 3) // base-script-region
	parts[0] = strings.ToLower(parts[0])
	if len(parts) >= 3 {
		parts[2] = strings.ToUpper(parts[2])
	} else if len(parts) == 2 {
		// Could be language-region or language-script
		if len(parts[1]) == 2 {
			parts[1] = strings.ToUpper(parts[1])
		} else {
			// Probably a script tag (4 chars) — title-case it manually
			if len(parts[1]) > 0 {
				parts[1] = strings.ToUpper(parts[1][:1]) + strings.ToLower(parts[1][1:])
			}
		}
	}
	return strings.Join(parts, "-")
}

// Negotiator selects the best UI locale from Accept-Language, cookie, and config defaults.
type Negotiator struct {
	registry *LocaleRegistry
}

// NewNegotiator creates a Negotiator backed by the given LocaleRegistry.
func NewNegotiator(reg *LocaleRegistry) *Negotiator {
	return &Negotiator{registry: reg}
}

// Negotiate resolves the best locale using priority:
// cookie.locale > Accept-Language > ui.default_locale > "en".
func (n *Negotiator) Negotiate(acceptLanguage, cookieLocale, defaultLocale string) string {
	// 1. Cookie locale
	if cookieLocale != "" {
		normalized := NormalizeLocaleTag(cookieLocale)
		if n.registry.IsSupported(normalized) {
			return normalized
		}
		// Fallback: language-only tag
		if langOnly := languageOnly(normalized); n.registry.IsSupported(langOnly) {
			return langOnly
		}
	}

	// 2. Accept-Language header
	if acceptLanguage != "" {
		tags := ParseAcceptLanguage(acceptLanguage)
		for _, tag := range tags {
			normalized := NormalizeLocaleTag(tag)
			if n.registry.IsSupported(normalized) {
				return normalized
			}
			// Fallback: language-only tag
			if langOnly := languageOnly(normalized); n.registry.IsSupported(langOnly) {
				return langOnly
			}
		}
	}

	// 3. Config default
	if defaultLocale != "" {
		normalized := NormalizeLocaleTag(defaultLocale)
		if n.registry.IsSupported(normalized) {
			return normalized
		}
	}

	// 4. Ultimate fallback
	return "en"
}

// ParseAcceptLanguage parses an Accept-Language header value into locale tags
// sorted by quality factor (q) descending.
func ParseAcceptLanguage(header string) []string {
	type qTag struct {
		tag string
		q   float64
	}

	parts := strings.Split(header, ",")
	tags := make([]qTag, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		q := 1.0
		tag := part
		if idx := strings.IndexByte(part, ';'); idx != -1 {
			tag = strings.TrimSpace(part[:idx])
			qPart := strings.TrimSpace(part[idx+1:])
			if strings.HasPrefix(qPart, "q=") {
				if parsed, err := strconv.ParseFloat(qPart[2:], 64); err == nil {
					q = parsed
				}
			}
		}
		tags = append(tags, qTag{tag: tag, q: q})
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].q > tags[j].q
	})

	result := make([]string, len(tags))
	for i, t := range tags {
		result[i] = t.tag
	}
	return result
}

// languageOnly extracts the primary language subtag from a locale tag.
func languageOnly(tag string) string {
	tag, _, _ = strings.Cut(tag, "-")
	tag, _, _ = strings.Cut(tag, "_")
	return tag
}
