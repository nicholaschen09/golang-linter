package bugs

import (
	"go/ast"
	"go/types"

	"github.com/nicholas/glint/pkg/rule"
)

type UncheckedError struct{}

func (UncheckedError) Name() string        { return "unchecked-error" }
func (UncheckedError) Category() rule.Category { return rule.CategoryBugs }
func (UncheckedError) Severity() rule.Severity { return rule.SeverityError }
func (UncheckedError) Description() string {
	return "Detects ignored error return values"
}
func (UncheckedError) NeedsTypeInfo() bool { return true }
func (UncheckedError) NodeTypes() []ast.Node {
	return []ast.Node{(*ast.ExprStmt)(nil)}
}

func (UncheckedError) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
	stmt, ok := node.(*ast.ExprStmt)
	if !ok || ctx.TypeInfo == nil {
		return nil
	}

	call, ok := stmt.X.(*ast.CallExpr)
	if !ok {
		return nil
	}

	t := ctx.TypeInfo.TypeOf(call)
	if t == nil {
		return nil
	}

	if hasErrorResult(t) {
		return []rule.Diagnostic{{
			Rule:     "unchecked-error",
			Category: rule.CategoryBugs,
			Severity: rule.SeverityError,
			Pos:      ctx.FileSet.Position(call.Pos()),
			End:      ctx.FileSet.Position(call.End()),
			Message:  "error return value is not checked",
		}}
	}

	return nil
}

func hasErrorResult(t types.Type) bool {
	switch typ := t.(type) {
	case *types.Named:
		return isErrorType(typ)
	case *types.Tuple:
		for i := 0; i < typ.Len(); i++ {
			if isErrorType(typ.At(i).Type()) {
				return true
			}
		}
	}
	return false
}

func isErrorType(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	for i := 0; i < iface.NumMethods(); i++ {
		if iface.Method(i).Name() == "Error" {
			sig, ok := iface.Method(i).Type().(*types.Signature)
			if ok && sig.Params().Len() == 0 && sig.Results().Len() == 1 {
				if basic, ok := sig.Results().At(0).Type().(*types.Basic); ok && basic.Kind() == types.String {
					return true
				}
			}
		}
	}
	return false
}

func init() {
	rule.Register(UncheckedError{})
}
