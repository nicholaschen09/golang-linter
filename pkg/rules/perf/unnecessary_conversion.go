package perf

import (
	"go/ast"
	"go/types"

	"github.com/nicholas/glint/pkg/rule"
)

type UnnecessaryConversion struct{}

func (UnnecessaryConversion) Name() string            { return "unnecessary-conversion" }
func (UnnecessaryConversion) Category() rule.Category { return rule.CategoryPerf }
func (UnnecessaryConversion) Severity() rule.Severity { return rule.SeverityWarning }
func (UnnecessaryConversion) Description() string {
	return "Detects redundant type conversions (e.g., int(x) where x is already int)"
}
func (UnnecessaryConversion) NeedsTypeInfo() bool { return true }
func (UnnecessaryConversion) NodeTypes() []ast.Node {
	return []ast.Node{(*ast.CallExpr)(nil)}
}

func (UnnecessaryConversion) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
	call, ok := node.(*ast.CallExpr)
	if !ok || ctx.TypeInfo == nil {
		return nil
	}

	if len(call.Args) != 1 {
		return nil
	}

	// Check if this is a type conversion (not a function call)
	tv, ok := ctx.TypeInfo.Types[call.Fun]
	if !ok || !tv.IsType() {
		return nil
	}

	argType := ctx.TypeInfo.TypeOf(call.Args[0])
	if argType == nil {
		return nil
	}

	convType := tv.Type
	if types.Identical(argType, convType) {
		return []rule.Diagnostic{{
			Rule:     "unnecessary-conversion",
			Category: rule.CategoryPerf,
			Severity: rule.SeverityWarning,
			Pos:      ctx.FileSet.Position(call.Pos()),
			End:      ctx.FileSet.Position(call.End()),
			Message:  "unnecessary type conversion; expression is already of type " + argType.String(),
		}}
	}

	return nil
}

func init() {
	rule.Register(UnnecessaryConversion{})
}
