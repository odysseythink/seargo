package builtin

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	if as := answerer.GlobalAnswerer(); as != nil {
		as.Register(newStatisticsAnswerer())
	}
}

type statisticsAnswerer struct {
	info answerer.AnswererInfo
}

func newStatisticsAnswerer() *statisticsAnswerer {
	return &statisticsAnswerer{
		info: answerer.AnswererInfo{
			Name:        "statistics",
			Description: "Compute statistics (min, max, avg, sum, prod, range) on a list of numbers",
			Keywords:    []string{"min", "max", "avg", "sum", "prod", "range"},
			Examples: []string{
				"avg 1 2 3",
				"sum 10 20 30",
				"min 5 3 8 1",
			},
		},
	}
}

func (a *statisticsAnswerer) Keywords() []string {
	return a.info.Keywords
}

func (a *statisticsAnswerer) Info() answerer.AnswererInfo {
	return a.info
}

func (a *statisticsAnswerer) Answer(ctx *answerer.AnswerContext) []models.Result {
	parts := strings.Fields(ctx.Query)
	if len(parts) < 2 {
		return nil
	}
	op := strings.ToLower(parts[0])
	nums := parseNumbers(parts[1:])
	if len(nums) == 0 {
		return nil
	}

	var result float64
	switch op {
	case "min":
		result = nums[0]
		for _, n := range nums[1:] {
			if n < result {
				result = n
			}
		}
	case "max":
		result = nums[0]
		for _, n := range nums[1:] {
			if n > result {
				result = n
			}
		}
	case "avg":
		sum := 0.0
		for _, n := range nums {
			sum += n
		}
		result = sum / float64(len(nums))
	case "sum":
		for _, n := range nums {
			result += n
		}
	case "prod":
		result = 1
		for _, n := range nums {
			result *= n
		}
	case "range":
		min, max := nums[0], nums[0]
		for _, n := range nums[1:] {
			if n < min {
				min = n
			}
			if n > max {
				max = n
			}
		}
		result = max - min
	default:
		return nil
	}

	// Format the value list
	valStrs := make([]string, len(nums))
	for i, n := range nums {
		valStrs[i] = formatNumber(n)
	}

	answer := fmt.Sprintf("%s(%s) = %s", op, strings.Join(valStrs, ", "), formatNumber(result))
	return []models.Result{{
		Kind:    "answer",
		Title:   answer,
		Content: fmt.Sprintf("Computed %s of %d number(s)", op, len(nums)),
		Engine:  "statistics",
	}}
}

// parseNumbers converts string tokens to float64, skipping non-numeric tokens.
func parseNumbers(strs []string) []float64 {
	var nums []float64
	for _, s := range strs {
		n, err := strconv.ParseFloat(s, 64)
		if err == nil && !math.IsNaN(n) {
			nums = append(nums, n)
		}
	}
	return nums
}

// formatNumber returns the number as an integer string if whole, or a trimmed decimal otherwise.
func formatNumber(f float64) string {
	if f == math.Trunc(f) && !math.IsInf(f, 0) {
		return fmt.Sprintf("%.0f", f)
	}
	s := fmt.Sprintf("%f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
