package query

import (
	"fmt"
	"strings"

	"github.com/seargo/seargo/pkg/models"
)

// RawTextQuery 是查询解析入口，管理解析器链。
type RawTextQuery struct {
	raw string
}

// NewRawTextQuery 创建解析器实例。
func NewRawTextQuery(raw string) *RawTextQuery {
	return &RawTextQuery{raw: raw}
}

// Parse 按 SearXNG 语法顺序解析原始查询：timeout → language → external_bang → bang → autocomplete。
// 未识别的语法片段保留为搜索词（graceful degradation）。
func (rtq *RawTextQuery) Parse(
	engineShortcuts map[string]string,
	engineNames []string,
	categories []models.Category,
	languages map[string]string,
) (*ParsedQuery, error) {

	timeoutP := TimeoutParser{}
	langP := LanguageParser{}
	extBangP := ExternalBangParser{}
	bangP := NewBangParser(engineShortcuts, engineNames, categories)
	autoP := AutocompleteTriggerParser{}

	tokens := strings.Fields(rtq.raw)

	var parts []QueryPart
	var userTerms []string
	var autocompleteTrigger bool

	for i, token := range tokens {
		// 1. Timeout
		if timeoutP.Check(token) {
			part, _ := timeoutP.Parse(token)
			parts = append(parts, part)
			continue
		}
		// 2. Language
		if langP.Check(token) {
			if part, ok := langP.Parse(token); ok {
				parts = append(parts, part)
				continue
			}
		}
		// 3. External bang (必须在 Bang 之前！)
		if extBangP.Check(token) {
			part, _ := extBangP.Parse(token)
			parts = append(parts, part)
			continue
		}
		// 4. Bang
		if bangP.Check(token) {
			part, _ := bangP.Parse(token)
			parts = append(parts, part)
			continue
		}
		// 5. Autocomplete trigger（仅第一个 token）
		if i == 0 {
			if autoP.Check(token, true) {
				autocompleteTrigger = true
				continue
			}
			// 处理 ?golang（无空格）：? 是 autocomplete 前缀
			if len(token) > 1 && token[0] == '?' {
				autocompleteTrigger = true
				userTerms = append(userTerms, token[1:])
				continue
			}
		}

		userTerms = append(userTerms, token)
	}

	return buildParsedQuery(rtq.raw, parts, userTerms, autocompleteTrigger), nil
}

func buildParsedQuery(raw string, parts []QueryPart, userTerms []string, autocompleteTrigger bool) *ParsedQuery {
	pq := &ParsedQuery{
		Terms:               userTerms,
		RawQuery:            raw,
		PageNo:              1,
		AutocompleteTrigger: autocompleteTrigger,
	}

	for _, part := range parts {
		switch part.Type {
		case PartTimeout:
			var secs float64
			fmt.Sscanf(part.Value, "%f", &secs)
			pq.Timeout = secs
		case PartLanguage:
			pq.Lang = part.Value
		case PartExternalBang:
			pq.ExternalBang = part.Value
		case PartBang:
			if part.Extra["kind"] == "engine" {
				pq.EngineRefs = appendUnique(pq.EngineRefs, part.Value)
			} else {
				pq.Categories = appendUniqueCats(pq.Categories, models.Category(part.Value))
			}
			pq.Specific = true
		}
	}

	return pq
}

func appendUnique[T comparable](slice []T, item T) []T {
	for _, v := range slice {
		if v == item {
			return slice
		}
	}
	return append(slice, item)
}

func appendUniqueCats(slice []models.Category, item models.Category) []models.Category {
	for _, v := range slice {
		if v == item {
			return slice
		}
	}
	return append(slice, item)
}
