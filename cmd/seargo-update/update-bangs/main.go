package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/seargo/seargo/cmd/seargo-update/internal"
)

const (
	defaultBangsURL = "https://duckduckgo.com/bang.js"
	leafKey         = "\x10"
	sepQuery        = "\x02"
	sepRank         = "\x01"
	httpsColon      = "https:"
	httpColon       = "http:"
)

type bangDefinition struct {
	URL  string `json:"u"`
	Rank int    `json:"r"`
	Tag  string `json:"t"`
}

type output struct {
	Version int                    `json:"version"`
	Trie    map[string]interface{} `json:"trie"`
}

func main() {
	var (
		out      = flag.String("out", "data/external_bangs.json", "output JSON path")
		bangsURL = flag.String("bangs-url", defaultBangsURL, "DuckDuckGo bang.js URL")
	)
	flag.Parse()

	if err := Run(*out, nil, *bangsURL); err != nil {
		fmt.Fprintf(os.Stderr, "update-bangs: %v\n", err)
		os.Exit(1)
	}
}

// Run fetches DuckDuckGo bang definitions and writes data/external_bangs.json.
func Run(outPath string, client fetch.Client, bangsURL string) error {
	h := fetch.New(client)
	ctx := context.Background()

	body, err := h.Get(ctx, bangsURL)
	if err != nil {
		return fmt.Errorf("fetch bangs: %w", err)
	}

	var definitions []bangDefinition
	if err := json.Unmarshal(body, &definitions); err != nil {
		return fmt.Errorf("parse bangs JSON: %w", err)
	}

	trie := make(map[string]interface{})
	bangUrls := make(map[string]string)

	for _, def := range definitions {
		bangURL := def.URL
		if !strings.Contains(bangURL, "{{{s}}}") {
			continue
		}
		bangURL = strings.ReplaceAll(bangURL, "{{{s}}}", sepQuery)

		if strings.HasPrefix(bangURL, httpsColon+"//") {
			bangURL = bangURL[len(httpsColon):]
		}

		var defStr string
		candidate := bangURL + sepRank + fmt.Sprintf("%d", def.Rank)
		if strings.HasPrefix(bangURL, httpColon+"//") {
			if existing, ok := bangUrls[bangURL[len(httpColon):]]; ok {
				defStr = existing
			} else if existing, ok := bangUrls[bangURL]; ok {
				defStr = existing
			} else {
				defStr = candidate
			}
		} else {
			if existing, ok := bangUrls[bangURL]; ok {
				defStr = existing
			} else {
				defStr = candidate
			}
		}
		bangUrls[bangURL] = defStr

		node := trie
		for i := 0; i < len(def.Tag); i++ {
			ch := string(def.Tag[i])
			next, ok := node[ch]
			if !ok {
				m := make(map[string]interface{})
				node[ch] = m
				node = m
				continue
			}
			node = next.(map[string]interface{})
		}
		node[leafKey] = defStr
	}

	mergeWhenNoLeaf(trie)
	optimizeLeaf(nil, "", trie)

	return writeJSON(outPath, output{Version: 0, Trie: trie})
}

// mergeWhenNoLeaf minimizes nodes by merging a child that has no leaf into its
// parent with a concatenated key.
func mergeWhenNoLeaf(node map[string]interface{}) {
	restart := false
	keys := make([]string, 0, len(node))
	for k := range node {
		keys = append(keys, k)
	}

	for _, key := range keys {
		if key == leafKey {
			continue
		}
		value, ok := node[key].(map[string]interface{})
		if !ok {
			continue
		}
		valueKeys := make([]string, 0, len(value))
		for k := range value {
			valueKeys = append(valueKeys, k)
		}

		_, hasLeaf := value[leafKey]
		if !hasLeaf {
			for _, valueKey := range valueKeys {
				node[key+valueKey] = value[valueKey]
				if child, ok := value[valueKey].(map[string]interface{}); ok {
					mergeWhenNoLeaf(child)
				}
			}
			delete(node, key)
			restart = true
		} else {
			mergeWhenNoLeaf(value)
		}
	}

	if restart {
		mergeWhenNoLeaf(node)
	}
}

// optimizeLeaf replaces single-leaf maps with the leaf string in their parent.
func optimizeLeaf(parent map[string]interface{}, parentKey string, node map[string]interface{}) {
	if len(node) == 1 {
		if leaf, ok := node[leafKey]; ok && parent != nil {
			parent[parentKey] = leaf
			return
		}
	}
	for key, value := range node {
		if child, ok := value.(map[string]interface{}); ok {
			optimizeLeaf(node, key, child)
		}
	}
}

func writeJSON(outPath string, v output) error {
	enc, err := json.MarshalIndent(toSorted(v), "", "    ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	enc = append(enc, '\n')

	tmp := outPath + ".tmp"
	if err := os.WriteFile(tmp, enc, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	return os.Rename(tmp, outPath)
}

func toSorted(v output) map[string]interface{} {
	return map[string]interface{}{
		"version": v.Version,
		"trie":    sortedMap(v.Trie),
	}
}

func sortedMap(m map[string]interface{}) map[string]interface{} {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]interface{}, len(keys))
	for _, k := range keys {
		ordered[k] = m[k]
	}
	return ordered
}
