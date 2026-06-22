package preferences

import (
	"strconv"
	"strings"
)

// validateField validates a single preference field against config choices.
// Returns the validated value, falling back to def when invalid or locked.
func validateField(key, value string, validChoices map[string][]string, locked map[string]bool, def string) string {
	if locked[key] {
		return def
	}
	choices, hasChoices := validChoices[key]
	if !hasChoices {
		// No constraints → basic type validation
		switch key {
		case "safesearch":
			return validateSafeSearch(value, 1)
		case "language":
			// Must be in configured languages list
			if choices, ok := validChoices["language"]; ok {
				for _, c := range choices {
					if c == value {
						return value
					}
				}
			}
			return def
		default:
			return value
		}
	}
	for _, c := range choices {
		if c == value {
			return value
		}
	}
	return def
}

// validateSafeSearch ensures safesearch is 0, 1, or 2; otherwise returns the default.
func validateSafeSearch(value string, def int) string {
	if value == "" {
		return ""
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 || n > 2 {
		return strconv.Itoa(def)
	}
	return value
}

// commaSplit splits a value by comma, trims whitespace, and removes empties.
func commaSplit(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// commaJoin joins values with comma.
func commaJoin(values []string) string {
	return strings.Join(values, ",")
}
