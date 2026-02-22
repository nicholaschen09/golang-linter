package security

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/nicholas/glint/pkg/rule"
)

var sqlFuncNames = map[string]bool{
	"Query":    true,
	"QueryRow": true,
	"Exec":     true,
	"Prepare":  true,
}

type SQLInjection struct{}

func (SQLInjection) Name() string            { return "sql-injection" }
func (SQLInjection) Category() rule.Category { return rule.CategorySecurity }
func (SQLInjection) Severity() rule.Severity { return rule.SeverityError }
func (SQLInjection) Description() string {
	return "Detects potential SQL injection via string concatenation in SQL query functions"
}
func (SQLInjection) NeedsTypeInfo() bool { return true }
func (SQLInjection) NodeTypes() []ast.Node {
	return []ast.Node{(*ast.CallExpr)(nil)}
}

func (SQLInjection) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
	call, ok := node.(*ast.CallExpr)
	if !ok || ctx.TypeInfo == nil {
		return nil
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	if !sqlFuncNames[sel.Sel.Name] {
		return nil
	}

	// Check if the receiver looks like a database type
	recvType := ctx.TypeInfo.TypeOf(sel.X)
	if recvType == nil || !looksLikeDBType(recvType) {
		return nil
	}

	if len(call.Args) == 0 {
		return nil
	}

	queryArg := call.Args[0]

	if containsStringConcat(queryArg) || containsSprintfCall(queryArg) {
		return []rule.Diagnostic{{
			Rule:     "sql-injection",
			Category: rule.CategorySecurity,
			Severity: rule.SeverityError,
			Pos:      ctx.FileSet.Position(queryArg.Pos()),
			End:      ctx.FileSet.Position(queryArg.End()),
			Message:  "potential SQL injection: use parameterized queries instead of string concatenation",
		}}
	}

	return nil
}

func looksLikeDBType(t types.Type) bool {
	name := t.String()
	return strings.Contains(name, "sql.DB") ||
		strings.Contains(name, "sql.Tx") ||
		strings.Contains(name, "sql.Conn") ||
		strings.Contains(name, "sqlx.DB") ||
		strings.Contains(name, "sqlx.Tx")
}

func containsStringConcat(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	return bin.Op.String() == "+"
}

func containsSprintfCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "fmt" && (sel.Sel.Name == "Sprintf" || sel.Sel.Name == "Sprint")
}

func init() {
	rule.Register(SQLInjection{})
}
