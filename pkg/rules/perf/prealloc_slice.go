package perf

import (
	"go/ast"
	"go/types"

	"github.com/nicholas/glint/pkg/rule"
)

type PreallocSlice struct{}

func (PreallocSlice) Name() string            { return "prealloc-slice" }
func (PreallocSlice) Category() rule.Category { return rule.CategoryPerf }
func (PreallocSlice) Severity() rule.Severity { return rule.SeverityWarning }
func (PreallocSlice) Description() string {
	return "Suggests preallocating slices that are grown inside loops with append"
}
func (PreallocSlice) NeedsTypeInfo() bool { return true }
func (PreallocSlice) NodeTypes() []ast.Node {
	return []ast.Node{
		(*ast.RangeStmt)(nil),
		(*ast.ForStmt)(nil),
	}
}

func (PreallocSlice) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
	if ctx.TypeInfo == nil {
		return nil
	}

	var body *ast.BlockStmt
	switch n := node.(type) {
	case *ast.RangeStmt:
		body = n.Body
	case *ast.ForStmt:
		body = n.Body
	default:
		return nil
	}

	if body == nil {
		return nil
	}

	var diags []rule.Diagnostic
	for _, stmt := range body.List {
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok || len(assign.Rhs) != 1 {
			continue
		}

		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			continue
		}

		fn, ok := call.Fun.(*ast.Ident)
		if !ok || fn.Name != "append" {
			continue
		}

		if len(call.Args) < 1 {
			continue
		}

		// Check the first arg is a slice being appended to
		t := ctx.TypeInfo.TypeOf(call.Args[0])
		if t == nil {
			continue
		}
		if _, ok := t.Underlying().(*types.Slice); !ok {
			continue
		}

		// Verify the LHS is the same variable as the first arg
		if len(assign.Lhs) == 1 {
			lhsIdent, lOk := assign.Lhs[0].(*ast.Ident)
			argIdent, rOk := call.Args[0].(*ast.Ident)
			if lOk && rOk && lhsIdent.Name == argIdent.Name {
				diags = append(diags, rule.Diagnostic{
					Rule:     "prealloc-slice",
					Category: rule.CategoryPerf,
					Severity: rule.SeverityWarning,
					Pos:      ctx.FileSet.Position(assign.Pos()),
					End:      ctx.FileSet.Position(assign.End()),
					Message:  "consider preallocating '" + lhsIdent.Name + "' with make([]T, 0, expectedLen)",
				})
			}
		}
	}

	return diags
}

func init() {
	rule.Register(PreallocSlice{})
}
