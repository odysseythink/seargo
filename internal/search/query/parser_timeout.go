package query

import (
	"fmt"
	"regexp"
	"strconv"
)

var timeoutRegex = regexp.MustCompile(`^<\d+$`)

// TimeoutParser 解析用户指定的超时语法，如 <3 表示 3 秒，<850 表示 0.85 秒。
type TimeoutParser struct{}

func (p TimeoutParser) Check(raw string) bool {
	return timeoutRegex.MatchString(raw)
}

func (p TimeoutParser) Parse(raw string) (QueryPart, bool) {
	if !p.Check(raw) {
		return QueryPart{}, false
	}
	n, err := strconv.Atoi(raw[1:])
	if err != nil {
		return QueryPart{}, false
	}
	seconds := float64(n)
	if n >= 100 {
		seconds = float64(n) / 1000.0
	}
	return QueryPart{
		Type:  PartTimeout,
		Value: fmt.Sprintf("%f", seconds),
	}, true
}
