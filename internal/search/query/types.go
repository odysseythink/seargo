package query

import "github.com/seargo/seargo/pkg/models"

// PartType 枚举查询语法片段的类型。
type PartType string

const (
	PartTimeout             PartType = "timeout"
	PartLanguage            PartType = "language"
	PartExternalBang        PartType = "external_bang"
	PartBang                PartType = "bang"
	PartAutocompleteTrigger PartType = "autocomplete_trigger"
)

// QueryPart 是单个解析器产生的语义片段。
type QueryPart struct {
	Type  PartType
	Value string
	Extra map[string]string
}

// Parser 是单类语法解析器接口。
type Parser interface {
	Check(raw string) bool
	Parse(raw string) (QueryPart, bool)
}

// ParsedQuery 是查询解析后的结构化结果。
type ParsedQuery struct {
	Terms              []string
	RawQuery           string
	EngineRefs         []string
	Categories         []models.Category
	Lang               string
	Timeout            float64 // 秒
	TimeRange          string
	PageNo             int
	SafeSearch         int
	ExternalBang       string
	AutocompleteTrigger bool
	Specific           bool
}
