package builtin

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
	"strings"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
)

type calculatorPlugin struct{}

func init() {
	plugin.RegisterBuiltin("calculator", func() plugin.Plugin { return &calculatorPlugin{} })
}

func (p *calculatorPlugin) ID() string { return "calculator" }

func (p *calculatorPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                "calculator",
		Name:              "Calculator",
		Description:       "Evaluate mathematical expressions",
		PreferenceSection: "query",
		Examples:          []string{"1+1", "2*3", "10/2", "2^3", "calc 1+1"},
	}
}

func (p *calculatorPlugin) Init(ctx *plugin.AppContext) bool                { return true }
func (p *calculatorPlugin) PreSearch(ctx *plugin.SearchContext) bool       { return true }
func (p *calculatorPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool { return true }

func (p *calculatorPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	if ctx.PageNo > 1 {
		return nil
	}

	q := strings.TrimSpace(ctx.Query)
	expr := q
	if strings.HasPrefix(strings.ToLower(q), "calc ") {
		expr = strings.TrimSpace(q[5:])
	}
	if expr == "" {
		return nil
	}

	val, err := evalExpr(expr)
	if err != nil {
		return nil
	}

	var resultStr string
	if val == math.Trunc(val) && !math.IsInf(val, 0) && !math.IsNaN(val) {
		resultStr = fmt.Sprintf("%s = %.0f", expr, val)
	} else {
		resultStr = fmt.Sprintf("%s = %g", expr, val)
	}

	return []models.Result{{
		Kind:    "answer",
		Title:   resultStr,
		Content: fmt.Sprintf("%g", val),
		Engine:  "calculator",
	}}
}

func evalExpr(input string) (float64, error) {
	expr, err := parser.ParseExpr(input)
	if err != nil {
		return 0, err
	}
	return evalAST(expr)
}

func evalAST(e ast.Expr) (float64, error) {
	switch n := e.(type) {
	case *ast.BasicLit:
		if n.Kind == token.INT || n.Kind == token.FLOAT {
			return strconv.ParseFloat(n.Value, 64)
		}
		return 0, fmt.Errorf("unexpected literal kind: %v", n.Kind)

	case *ast.ParenExpr:
		return evalAST(n.X)

	case *ast.UnaryExpr:
		v, err := evalAST(n.X)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.SUB:
			return -v, nil
		case token.ADD:
			return v, nil
		}
		return 0, fmt.Errorf("unexpected unary operator: %v", n.Op)

	case *ast.BinaryExpr:
		left, err := evalAST(n.X)
		if err != nil {
			return 0, err
		}
		right, err := evalAST(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			return left / right, nil
		case token.XOR:
			return math.Pow(left, right), nil
		}
		return 0, fmt.Errorf("unexpected binary operator: %v", n.Op)

	default:
		return 0, fmt.Errorf("unexpected expression type: %T", e)
	}
}
