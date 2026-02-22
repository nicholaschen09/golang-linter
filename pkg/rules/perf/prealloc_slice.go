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
func (PreallocSlice) NeedsTypeInfo() bool  { return true }
func (PreallocSlice) NodeTypes() []ast.Node { return nil }

func (PreallocSlice) Check(_ *rule.Context, _ ast.Node) []rule.Diagnostic {
	return nil
}

// CheckFile walks the file to find append-in-loop patterns while tracking
// which slice variables were already preallocated with make().
func (PreallocSlice) CheckFile(ctx *rule.Context) []rule.Diagnostic {
	if ctx.TypeInfo == nil {
		return nil
	}

	var diags []rule.Diagnostic

	ast.Inspect(ctx.File, func(n ast.Node) bool {
		fn, fnOk := n.(*ast.FuncDecl)
		if !fnOk || fn.Body == nil {
			return true
		}
		diags = append(diags, checkFuncBody(ctx, fn.Body)...)
		return false
	})

	return diags
}

func checkFuncBody(ctx *rule.Context, body *ast.BlockStmt) []rule.Diagnostic {
	diags := make([]rule.Diagnostic, 0)
	preallocated := make(map[string]bool)

	for _, stmt := range body.List {
		trackMakeAllocations(stmt, preallocated)
		diags = append(diags, checkStmtForAppendInLoop(ctx, stmt, preallocated)...)
	}

	return diags
}

func trackMakeAllocations(stmt ast.Stmt, preallocated map[string]bool) {
	assign, assignOk := stmt.(*ast.AssignStmt)
	if !assignOk || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return
	}

	ident, identOk := assign.Lhs[0].(*ast.Ident)
	if !identOk {
		return
	}

	call, callOk := assign.Rhs[0].(*ast.CallExpr)
	if !callOk {
		return
	}

	fnIdent, fnOk := call.Fun.(*ast.Ident)
	if !fnOk || fnIdent.Name != "make" {
		return
	}

	// make([]T, len) or make([]T, len, cap) — both count as preallocated
	if len(call.Args) >= 2 {
		preallocated[ident.Name] = true
	}
}

func checkStmtForAppendInLoop(
	ctx *rule.Context,
	stmt ast.Stmt,
	preallocated map[string]bool,
) []rule.Diagnostic {
	var body *ast.BlockStmt

	switch s := stmt.(type) {
	case *ast.RangeStmt:
		body = s.Body
	case *ast.ForStmt:
		body = s.Body
	default:
		return nil
	}

	if body == nil {
		return nil
	}

	var diags []rule.Diagnostic
	for _, loopStmt := range body.List {
		assign, assignOk := loopStmt.(*ast.AssignStmt)
		if !assignOk || len(assign.Rhs) != 1 {
			continue
		}

		call, callOk := assign.Rhs[0].(*ast.CallExpr)
		if !callOk {
			continue
		}

		fnIdent, fnOk := call.Fun.(*ast.Ident)
		if !fnOk || fnIdent.Name != "append" || len(call.Args) < 1 {
			continue
		}

		t := ctx.TypeInfo.TypeOf(call.Args[0])
		if t == nil {
			continue
		}
		if _, sliceOk := t.Underlying().(*types.Slice); !sliceOk {
			continue
		}

		if len(assign.Lhs) == 1 {
			lhsIdent, lOk := assign.Lhs[0].(*ast.Ident)
			argIdent, rOk := call.Args[0].(*ast.Ident)
			if lOk && rOk && lhsIdent.Name == argIdent.Name {
				if preallocated[lhsIdent.Name] {
					continue
				}
				diags = append(diags, rule.Diagnostic{
					Rule:     "prealloc-slice",
					Category: rule.CategoryPerf,
					Severity: rule.SeverityWarning,
					Pos:      ctx.FileSet.Position(assign.Pos()),
					End:      ctx.FileSet.Position(assign.End()),
					Message:  "consider preallocating '" + lhsIdent.Name + "'",
				})
			}
		}
	}

	return diags
}

func init() {
	rule.Register(PreallocSlice{})
}
