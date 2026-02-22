package security

import (
	"go/ast"
	"strings"

	"github.com/nicholas/glint/pkg/rule"
)

var defaultSecretPatterns = []string{
	"password", "passwd", "secret", "api_key", "apikey",
	"access_token", "auth_token", "private_key", "token",
}

type HardcodedSecret struct{}

func (HardcodedSecret) Name() string            { return "hardcoded-secret" }
func (HardcodedSecret) Category() rule.Category { return rule.CategorySecurity }
func (HardcodedSecret) Severity() rule.Severity { return rule.SeverityError }
func (HardcodedSecret) Description() string {
	return "Detects hardcoded secrets in string assignments (passwords, API keys, tokens)"
}
func (HardcodedSecret) NeedsTypeInfo() bool { return false }
func (HardcodedSecret) NodeTypes() []ast.Node {
	return []ast.Node{
		(*ast.AssignStmt)(nil),
		(*ast.ValueSpec)(nil),
		(*ast.KeyValueExpr)(nil),
	}
}

func (HardcodedSecret) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
	if a, aOk := node.(*ast.AssignStmt); aOk {
		return checkAssign(ctx, a)
	}
	if v, vOk := node.(*ast.ValueSpec); vOk {
		return checkValueSpec(ctx, v)
	}
	if kv, kvOk := node.(*ast.KeyValueExpr); kvOk {
		return checkKeyValue(ctx, kv)
	}
	return nil
}

func checkAssign(ctx *rule.Context, assign *ast.AssignStmt) []rule.Diagnostic {
	var diags []rule.Diagnostic
	for i, lhs := range assign.Lhs {
		ident, identOk := lhs.(*ast.Ident)
		if !identOk {
			continue
		}
		if i >= len(assign.Rhs) {
			continue
		}
		if isSecretName(ident.Name) && isStringLiteral(assign.Rhs[i]) {
			diags = append(diags, makeDiag(ctx, ident, ident.Name))
		}
	}
	return diags
}

func checkValueSpec(ctx *rule.Context, vs *ast.ValueSpec) []rule.Diagnostic {
	var diags []rule.Diagnostic
	for i, name := range vs.Names {
		if isSecretName(name.Name) && i < len(vs.Values) && isStringLiteral(vs.Values[i]) {
			diags = append(diags, makeDiag(ctx, name, name.Name))
		}
	}
	return diags
}

func checkKeyValue(ctx *rule.Context, kv *ast.KeyValueExpr) []rule.Diagnostic {
	key, keyIsLit := kv.Key.(*ast.BasicLit)
	if !keyIsLit {
		ident, identOk := kv.Key.(*ast.Ident)
		if identOk && isSecretName(ident.Name) && isStringLiteral(kv.Value) {
			return []rule.Diagnostic{makeDiag(ctx, ident, ident.Name)}
		}
		return nil
	}
	keyStr := strings.Trim(key.Value, `"`)
	if isSecretName(keyStr) && isStringLiteral(kv.Value) {
		return []rule.Diagnostic{{
			Rule:     "hardcoded-secret",
			Category: rule.CategorySecurity,
			Severity: rule.SeverityError,
			Pos:      ctx.FileSet.Position(kv.Pos()),
			End:      ctx.FileSet.Position(kv.End()),
			Message:  "potential hardcoded secret in key '" + keyStr + "'",
		}}
	}
	return nil
}

func makeDiag(ctx *rule.Context, ident *ast.Ident, name string) rule.Diagnostic {
	return rule.Diagnostic{
		Rule:     "hardcoded-secret",
		Category: rule.CategorySecurity,
		Severity: rule.SeverityError,
		Pos:      ctx.FileSet.Position(ident.Pos()),
		End:      ctx.FileSet.Position(ident.End()),
		Message:  "potential hardcoded secret in variable '" + name + "'",
	}
}

func isSecretName(name string) bool {
	lower := strings.ToLower(name)
	for _, pat := range defaultSecretPatterns {
		if strings.Contains(lower, pat) {
			return true
		}
	}
	return false
}

func isStringLiteral(expr ast.Expr) bool {
	lit, litOk := expr.(*ast.BasicLit)
	if !litOk {
		return false
	}
	val := strings.Trim(lit.Value, `"`+"`")
	return len(val) > 0 && val != "" && val != "''"
}

func init() {
	rule.Register(HardcodedSecret{})
}
