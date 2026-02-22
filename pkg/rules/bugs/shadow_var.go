package bugs

import (
	"go/ast"
	"go/token"

	"github.com/nicholas/glint/pkg/rule"
)

type ShadowVar struct{}

func (ShadowVar) Name() string            { return "shadow-var" }
func (ShadowVar) Category() rule.Category { return rule.CategoryBugs }
func (ShadowVar) Severity() rule.Severity { return rule.SeverityWarning }
func (ShadowVar) Description() string {
	return "Detects variable shadowing in inner scopes"
}
func (ShadowVar) NeedsTypeInfo() bool { return true }
func (ShadowVar) NodeTypes() []ast.Node {
	return []ast.Node{(*ast.AssignStmt)(nil)}
}

func (ShadowVar) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
	assign, isAssign := node.(*ast.AssignStmt)
	if !isAssign || assign.Tok != token.DEFINE {
		return nil
	}
	if ctx.TypeInfo == nil {
		return nil
	}

	var diags []rule.Diagnostic
	for _, lhs := range assign.Lhs {
		ident, identOk := lhs.(*ast.Ident)
		if !identOk || ident.Name == "_" {
			continue
		}

		obj := ctx.TypeInfo.ObjectOf(ident)
		if obj == nil {
			continue
		}

		scope := obj.Parent()
		if scope == nil {
			continue
		}

		for outer := scope.Parent(); outer != nil; outer = outer.Parent() {
			if shadowed := outer.Lookup(ident.Name); shadowed != nil {
				diags = append(diags, rule.Diagnostic{
					Rule:     "shadow-var",
					Category: rule.CategoryBugs,
					Severity: rule.SeverityWarning,
					Pos:      ctx.FileSet.Position(ident.Pos()),
					End:      ctx.FileSet.Position(ident.End()),
					Message:  "variable '" + ident.Name + "' shadows declaration at " + ctx.FileSet.Position(shadowed.Pos()).String(),
				})
				break
			}
		}
	}
	return diags
}

func init() {
	rule.Register(ShadowVar{})
}
