package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/stretchr/testify/assert"
)

func TestStatisticsAnswerer_Avg(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "avg 1 2 3"})
	assert.Len(t, results, 1)
	assert.Equal(t, "avg(1, 2, 3) = 2", results[0].Title)
	assert.Equal(t, "statistics", results[0].Engine)
}

func TestStatisticsAnswerer_Avg_Float(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "avg 1.5 2.5 3.5"})
	assert.Len(t, results, 1)
	assert.Equal(t, "avg(1.5, 2.5, 3.5) = 2.5", results[0].Title)
}

func TestStatisticsAnswerer_Sum(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "sum 1 2 3"})
	assert.Len(t, results, 1)
	assert.Equal(t, "sum(1, 2, 3) = 6", results[0].Title)
}

func TestStatisticsAnswerer_Min(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "min 3 1 2"})
	assert.Len(t, results, 1)
	assert.Equal(t, "min(3, 1, 2) = 1", results[0].Title)
}

func TestStatisticsAnswerer_Max(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "max 3 1 2"})
	assert.Len(t, results, 1)
	assert.Equal(t, "max(3, 1, 2) = 3", results[0].Title)
}

func TestStatisticsAnswerer_Range(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "range 3 1 4"})
	assert.Len(t, results, 1)
	assert.Equal(t, "range(3, 1, 4) = 3", results[0].Title)
}

func TestStatisticsAnswerer_Prod(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "prod 2 3 4"})
	assert.Len(t, results, 1)
	assert.Equal(t, "prod(2, 3, 4) = 24", results[0].Title)
}

func TestStatisticsAnswerer_SingleNumber(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "avg 5"})
	assert.Len(t, results, 1)
	assert.Equal(t, "avg(5) = 5", results[0].Title)
}

func TestStatisticsAnswerer_NoNumbers(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	// Only the keyword, no numbers
	results := as.Ask(&answerer.AnswerContext{Query: "avg"})
	assert.Nil(t, results)
}

func TestStatisticsAnswerer_SkipsNonNumeric(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "sum 1 a 2 b 3"})
	assert.Len(t, results, 1)
	assert.Equal(t, "sum(1, 2, 3) = 6", results[0].Title)
}

func TestStatisticsAnswerer_NoMatch(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newStatisticsAnswerer()
	as.Register(a)

	// Not a matching keyword
	results := as.Ask(&answerer.AnswerContext{Query: "foo 1 2"})
	assert.Nil(t, results)
}

func TestStatisticsAnswerer_Keywords(t *testing.T) {
	a := newStatisticsAnswerer()
	kw := a.Keywords()
	assert.ElementsMatch(t, []string{"min", "max", "avg", "sum", "prod", "range"}, kw)
}

func TestStatisticsAnswerer_Info(t *testing.T) {
	a := newStatisticsAnswerer()
	info := a.Info()
	assert.Equal(t, "statistics", info.Name)
	assert.NotEmpty(t, info.Description)
	assert.NotEmpty(t, info.Examples)
}

func TestParseNumbers(t *testing.T) {
	nums := parseNumbers([]string{"1", "2", "3"})
	assert.Equal(t, []float64{1, 2, 3}, nums)
}

func TestParseNumbers_SkipsInvalid(t *testing.T) {
	nums := parseNumbers([]string{"1", "abc", "3.5", "", "x"})
	assert.Equal(t, []float64{1, 3.5}, nums)
}

func TestParseNumbers_Empty(t *testing.T) {
	nums := parseNumbers([]string{})
	assert.Empty(t, nums)
}

func TestFormatNumber_Whole(t *testing.T) {
	assert.Equal(t, "2", formatNumber(2.0))
	assert.Equal(t, "0", formatNumber(0.0))
	assert.Equal(t, "100", formatNumber(100.0))
}

func TestFormatNumber_Float(t *testing.T) {
	assert.Equal(t, "2.5", formatNumber(2.5))
	assert.Equal(t, "0.333333", formatNumber(0.333333))
}
