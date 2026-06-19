package query

import (
	"strings"

	"github.com/seargo/seargo/pkg/models"
)

// normalizeBang 规范化 bang 字符串：替换 -/_ 为空格，小写。
func normalizeBang(bang string) string {
	bang = strings.ReplaceAll(bang, "-", " ")
	bang = strings.ReplaceAll(bang, "_", " ")
	return strings.ToLower(bang)
}

// ExternalBangParser 解析 !!bang 语法，用于重定向到外部搜索引擎。
type ExternalBangParser struct{}

func (p ExternalBangParser) Check(raw string) bool {
	return len(raw) > 2 && raw[0] == '!' && raw[1] == '!' && raw[2] != '!'
}

func (p ExternalBangParser) Parse(raw string) (QueryPart, bool) {
	if !p.Check(raw) {
		return QueryPart{}, false
	}
	bang := normalizeBang(raw[2:])
	return QueryPart{
		Type:  PartExternalBang,
		Value: bang,
	}, true
}

// BangParser 解析 !bang 语法，匹配引擎 shortcut、引擎名或分类名。
type BangParser struct {
	shortcuts  map[string]string // shortcut → engine name
	names      []string
	categories []models.Category
}

// NewBangParser 创建 BangParser。
func NewBangParser(shortcuts map[string]string, names []string, categories []models.Category) BangParser {
	return BangParser{
		shortcuts:  shortcuts,
		names:      names,
		categories: categories,
	}
}

func (p BangParser) Check(raw string) bool {
	if len(raw) < 2 || raw[0] != '!' {
		return false
	}
	// 排除 external bang
	if len(raw) > 2 && raw[1] == '!' {
		return false
	}

	bang := normalizeBang(raw[1:])

	// 检查 shortcut
	if _, ok := p.shortcuts[bang]; ok {
		return true
	}
	// 检查引擎名
	for _, n := range p.names {
		if strings.EqualFold(n, bang) {
			return true
		}
	}
	// 检查分类名
	for _, c := range p.categories {
		if strings.EqualFold(string(c), bang) {
			return true
		}
	}
	return false
}

func (p BangParser) Parse(raw string) (QueryPart, bool) {
	if !p.Check(raw) {
		return QueryPart{}, false
	}

	bang := normalizeBang(raw[1:])

	// 1. 引擎 shortcut
	if name, ok := p.shortcuts[bang]; ok {
		return QueryPart{
			Type:  PartBang,
			Value: name,
			Extra: map[string]string{"kind": "engine"},
		}, true
	}

	// 2. 引擎名
	for _, n := range p.names {
		if strings.EqualFold(n, bang) {
			return QueryPart{
				Type:  PartBang,
				Value: n,
				Extra: map[string]string{"kind": "engine"},
			}, true
		}
	}

	// 3. 分类名
	for _, c := range p.categories {
		if strings.EqualFold(string(c), bang) {
			return QueryPart{
				Type:  PartBang,
				Value: string(c),
				Extra: map[string]string{"kind": "category"},
			}, true
		}
	}

	return QueryPart{}, false
}
