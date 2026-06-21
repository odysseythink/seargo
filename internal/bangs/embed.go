package bangs

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed external_bangs.json
var bangsData []byte

type rawTrie struct {
	Trie    map[string]interface{} `json:"trie"`
	Version int                    `json:"version"`
}

func parseTrie() (map[string]interface{}, error) {
	if len(bangsData) == 0 {
		return nil, fmt.Errorf("bangs: external_bangs.json is empty or not embedded")
	}
	var raw rawTrie
	if err := json.Unmarshal(bangsData, &raw); err != nil {
		return nil, fmt.Errorf("bangs: failed to parse external_bangs.json: %w", err)
	}
	if raw.Trie == nil || len(raw.Trie) == 0 {
		return nil, fmt.Errorf("bangs: external_bangs.json has empty trie")
	}
	return raw.Trie, nil
}
