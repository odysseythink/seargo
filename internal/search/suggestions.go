package search

import "strings"

// mergeSuggestions concatenates multiple suggestion lists, deduplicates
// case-insensitively (keeping first occurrence), and limits to 10 results.
func mergeSuggestions(allSuggestions [][]string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, suggestions := range allSuggestions {
		for _, s := range suggestions {
			lower := strings.ToLower(s)
			if seen[lower] {
				continue
			}
			seen[lower] = true
			result = append(result, s)
			if len(result) >= 10 {
				return result
			}
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}
