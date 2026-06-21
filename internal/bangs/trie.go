package bangs

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const (
	sepQuery = '\x02'
	sepRank  = '\x01'
	leafKey  = "\x10"
)

type BangDefinition struct {
	URL  string
	Rank int
}

type BangTrie struct {
	root map[string]interface{}
}

func NewBangTrie() (*BangTrie, error) {
	root, err := parseTrie()
	if err != nil {
		return nil, err
	}
	return &BangTrie{root: root}, nil
}

func (bt *BangTrie) Resolve(bang string, query string) *string {
	def := bt.resolveDef(bang)
	if def == nil {
		return nil
	}
	template := def.URL
	if query == "" {
		return bt.rootDomain(template)
	}
	result := strings.Replace(template, string(sepQuery), url.QueryEscape(query), 1)
	return &result
}

func (bt *BangTrie) Suggest(prefix string) []string {
	node := bt.walk(prefix)
	if node == nil {
		return nil
	}
	suggestions := bt.collectSubtree(prefix, node)
	sort.Strings(suggestions)
	return suggestions
}

func (bt *BangTrie) resolveDef(bang string) *BangDefinition {
	node := bt.root
	var lastDef *BangDefinition

	for i := 0; i < len(bang); i++ {
		next, ok := node[string(bang[i])]
		if !ok {
			return lastDef
		}
		switch v := next.(type) {
		case string:
			def := parseDef(v)
			lastDef = &def
			node = nil
		case map[string]interface{}:
			node = v
			if self, ok := node[leafKey]; ok {
				if s, ok := self.(string); ok {
					def := parseDef(s)
					lastDef = &def
				}
			}
		default:
			return lastDef
		}
	}

	if s, ok := node[leafKey].(string); ok {
		def := parseDef(s)
		return &def
	}
	if lastDef != nil {
		return lastDef
	}
	return nil
}

func (bt *BangTrie) walk(prefix string) map[string]interface{} {
	node := bt.root
	for i := 0; i < len(prefix); i++ {
		next, ok := node[string(prefix[i])]
		if !ok {
			return nil
		}
		switch v := next.(type) {
		case string:
			return nil
		case map[string]interface{}:
			node = v
		default:
			return nil
		}
	}
	return node
}

func (bt *BangTrie) collectSubtree(base string, node map[string]interface{}) []string {
	var results []string
	for k, v := range node {
		if k == leafKey {
			if _, ok := v.(string); ok && base != "" {
				results = append(results, base)
			}
			continue
		}
		childPath := base + k
		switch child := v.(type) {
		case string:
			results = append(results, childPath)
		case map[string]interface{}:
			results = append(results, bt.collectSubtree(childPath, child)...)
		}
	}
	return results
}

func parseDef(raw string) BangDefinition {
	parts := strings.SplitN(raw, string(sepRank), 2)
	if len(parts) != 2 {
		return BangDefinition{URL: raw, Rank: 0}
	}
	rank := 0
	fmt.Sscanf(parts[1], "%d", &rank)
	return BangDefinition{URL: parts[0], Rank: rank}
}

func (bt *BangTrie) rootDomain(template string) *string {
	t := template
	// Strip control characters used as sentinels before parsing
	t = strings.ReplaceAll(t, string(sepQuery), "")
	t = strings.ReplaceAll(t, string(sepRank), "")
	if strings.HasPrefix(t, "//") {
		t = "https:" + t
	}
	parsed, err := url.Parse(t)
	if err != nil {
		return nil
	}
	parsed.Path = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	result := parsed.String()
	if !strings.HasSuffix(result, "/") {
		result += "/"
	}
	return &result
}
