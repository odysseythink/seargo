package builtin

import (
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/stretchr/testify/assert"
)

func TestCalculatorPlugin_SimpleMath(t *testing.T) {
	p := &calculatorPlugin{}
	ctx := &plugin.SearchContext{
		Query:  "calc 1+1",
		PageNo: 1,
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "1+1 = 2", results[0].Title)
	assert.Equal(t, "calculator", results[0].Engine)
}

func TestCalculatorPlugin_Multiplication(t *testing.T) {
	p := &calculatorPlugin{}
	ctx := &plugin.SearchContext{
		Query:  "calc 2*3",
		PageNo: 1,
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "2*3 = 6", results[0].Title)
}

func TestCalculatorPlugin_Division(t *testing.T) {
	p := &calculatorPlugin{}
	ctx := &plugin.SearchContext{
		Query:  "calc 10/2",
		PageNo: 1,
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "10/2 = 5", results[0].Title)
}

func TestCalculatorPlugin_Power(t *testing.T) {
	p := &calculatorPlugin{}
	ctx := &plugin.SearchContext{
		Query:  "calc 2^3",
		PageNo: 1,
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "2^3 = 8", results[0].Title)
}

func TestCalculatorPlugin_Parens(t *testing.T) {
	p := &calculatorPlugin{}
	ctx := &plugin.SearchContext{
		Query:  "calc (1+2)*3",
		PageNo: 1,
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "(1+2)*3 = 9", results[0].Title)
}

func TestCalculatorPlugin_NoMatch(t *testing.T) {
	p := &calculatorPlugin{}
	ctx := &plugin.SearchContext{
		Query:  "regular search query",
		PageNo: 1,
	}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestCalculatorPlugin_PageGreaterThanOne(t *testing.T) {
	p := &calculatorPlugin{}
	ctx := &plugin.SearchContext{
		Query:  "calc 1+1",
		PageNo: 2,
	}
	results := p.PostSearch(ctx)
	assert.Empty(t, results)
}

func TestCalculatorPlugin_CaseInsensitive(t *testing.T) {
	p := &calculatorPlugin{}
	ctx := &plugin.SearchContext{
		Query:  "CALC 5+5",
		PageNo: 1,
	}
	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "5+5 = 10", results[0].Title)
}
