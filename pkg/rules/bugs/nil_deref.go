package bugs

import (
	"go/ast"
	"go/types"

	"github.com/nicholas/glint/pkg/rule"
)

type NilDeref struct{}

func (NilDeref) Name() string            { return "nil-deref" }
func (NilDeref) Category() rule.Category { return rule.CategoryBugs }
func (NilDeref) Severity() rule.Severity { return rule.SeverityError }
func (NilDeref) Description() string {
	return "Detects potential nil pointer dereferences after type assertions or map lookups without ok check"
}
func (NilDeref) NeedsTypeInfo() bool { return true }
func (NilDeref) NodeTypes() []ast.Node {
	return []ast.Node{(*ast.AssignStmt)(nil)}
}

func (NilDeref) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
	assign, ok := node.(*ast.AssignStmt)
	if !ok || ctx.TypeInfo == nil {
		return nil
	}

	if len(assign.Rhs) != 1 {
		return nil
	}

	var diags []rule.Diagnostic

	switch rhs := assign.Rhs[0].(type) {
	case *ast.TypeAssertExpr:
		// Single-value type assertion: v = x.(T) panics on nil
		if len(assign.Lhs) == 1 {
			t := ctx.TypeInfo.TypeOf(rhs.X)
			if t != nil && isNillable(t) {
				diags = append(diags, rule.Diagnostic{
					Rule:     "nil-deref",
					Category: rule.CategoryBugs,
					Severity: rule.SeverityError,
					Pos:      ctx.FileSet.Position(rhs.Pos()),
					End:      ctx.FileSet.Position(rhs.End()),
					Message:  "type assertion without ok check; will panic if value is nil or wrong type",
				})
			}
		}
	case *ast.IndexExpr:
		// Single-value map lookup: v = m[k] — if result is pointer type,
		// using it without ok check risks nil deref.
		if len(assign.Lhs) == 1 {
			t := ctx.TypeInfo.TypeOf(rhs.X)
			if t != nil {
				if mt, ok := t.Underlying().(*types.Map); ok {
					if isNillable(mt.Elem()) {
						diags = append(diags, rule.Diagnostic{
							Rule:     "nil-deref",
							Category: rule.CategoryBugs,
							Severity: rule.SeverityWarning,
							Pos:      ctx.FileSet.Position(rhs.Pos()),
							End:      ctx.FileSet.Position(rhs.End()),
							Message:  "map lookup of pointer type without ok check; zero value is nil",
						})
					}
				}
			}
		}
	}

	return diags
}

func isNillable(t types.Type) bool {
	switch t.Underlying().(type) {
	case *types.Pointer, *types.Interface, *types.Slice, *types.Map, *types.Chan, *types.Signature:
		return true
	}
	return false
}

func init() {
	rule.Register(NilDeref{})
}
