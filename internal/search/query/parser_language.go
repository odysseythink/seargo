package query

import (
	"strings"
	"unicode"
)

// languageNameToCode maps common language names to canonical language codes.
var languageNameToCode = map[string]string{
	"english":    "en",
	"chinese":    "zh",
	"german":     "de",
	"french":     "fr",
	"spanish":    "es",
	"japanese":   "ja",
	"korean":     "ko",
	"russian":    "ru",
	"italian":    "it",
	"portuguese": "pt",
	"arabic":     "ar",
	"dutch":      "nl",
}

// LanguageParser 解析语言指定语法 :en、:zh-CN、:english 等。
type LanguageParser struct{}

func (p LanguageParser) Check(raw string) bool {
	return len(raw) > 1 && raw[0] == ':'
}

func (p LanguageParser) Parse(raw string) (QueryPart, bool) {
	if !p.Check(raw) {
		return QueryPart{}, false
	}
	candidate := raw[1:]
	// 将下划线替换为横线（兼容 :en_us 写法）
	candidate = strings.ReplaceAll(candidate, "_", "-")
	candidate = strings.ToLower(candidate)

	// 1. 匹配内置语言名映射（优先于通用代码检查，避免 :english 误判为代码）
	if code, ok := languageNameToCode[candidate]; ok {
		return QueryPart{
			Type:  PartLanguage,
			Value: code,
		}, true
	}

	// 2. 直接匹配语言代码
	if isValidLanguageCode(candidate) {
		return QueryPart{
			Type:  PartLanguage,
			Value: normalizeLanguageCode(candidate),
		}, true
	}

	return QueryPart{}, false
}

// isValidLanguageCode 校验语言代码格式：2 字母，或 2 字母 + - + 2 字母地区码。
func isValidLanguageCode(code string) bool {
	if len(code) == 2 {
		return isAlpha(code)
	}
	if len(code) == 5 && code[2] == '-' {
		return isAlpha(code[:2]) && isAlpha(code[3:])
	}
	return false
}

func isAlpha(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return len(s) > 0
}

// normalizeLanguageCode 规范化大小写：基础码小写，地区码大写（如 zh-cn → zh-CN）。
func normalizeLanguageCode(code string) string {
	parts := strings.SplitN(code, "-", 2)
	parts[0] = strings.ToLower(parts[0])
	if len(parts) == 2 {
		parts[1] = strings.ToUpper(parts[1])
	}
	return strings.Join(parts, "-")
}
