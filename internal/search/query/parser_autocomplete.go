package query

// AutocompleteTriggerParser 检测查询是否以 ? 开头，触发 autocomplete 模式。
type AutocompleteTriggerParser struct{}

// Check 仅在 isFirstToken 为 true 且 raw == "?" 时返回 true。
func (p AutocompleteTriggerParser) Check(raw string, isFirstToken bool) bool {
	return isFirstToken && raw == "?"
}

func (p AutocompleteTriggerParser) Parse(raw string) (QueryPart, bool) {
	return QueryPart{
		Type:  PartAutocompleteTrigger,
		Value: "",
	}, true
}
