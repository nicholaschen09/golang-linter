package style

import (
	"go/ast"
	"strings"
	"unicode"

	"github.com/nicholas/glint/pkg/rule"
)

type NamingConvention struct{}

func (NamingConvention) Name() string            { return "naming-convention" }
func (NamingConvention) Category() rule.Category { return rule.CategoryStyle }
func (NamingConvention) Severity() rule.Severity { return rule.SeverityWarning }
func (NamingConvention) Description() string {
	return "Enforces Go naming conventions (MixedCaps, no underscores in exported names)"
}
func (NamingConvention) NeedsTypeInfo() bool { return false }
func (NamingConvention) NodeTypes() []ast.Node {
	return []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.TypeSpec)(nil),
		(*ast.ValueSpec)(nil),
	}
}

func (NamingConvention) Check(ctx *rule.Context, node ast.Node) []rule.Diagnostic {
	var diags []rule.Diagnostic

	if fd, fdOk := node.(*ast.FuncDecl); fdOk {
		if d := checkName(ctx, fd.Name); d != nil {
			diags = append(diags, *d)
		}
	}
	if ts, tsOk := node.(*ast.TypeSpec); tsOk {
		if d := checkName(ctx, ts.Name); d != nil {
			diags = append(diags, *d)
		}
	}
	if vs, vsOk := node.(*ast.ValueSpec); vsOk {
		for _, name := range vs.Names {
			if d := checkName(ctx, name); d != nil {
				diags = append(diags, *d)
			}
		}
	}

	return diags
}

func checkName(ctx *rule.Context, ident *ast.Ident) *rule.Diagnostic {
	if ident == nil || ident.Name == "_" || ident.Name == "main" || ident.Name == "init" {
		return nil
	}

	name := ident.Name
	if !ident.IsExported() {
		return nil
	}

	if strings.Contains(name, "_") {
		if name == strings.ToUpper(name) {
			return nil
		}
		d := rule.Diagnostic{
			Rule:     "naming-convention",
			Category: rule.CategoryStyle,
			Severity: rule.SeverityWarning,
			Pos:      ctx.FileSet.Position(ident.Pos()),
			End:      ctx.FileSet.Position(ident.End()),
			Message:  "exported name '" + name + "' should not contain underscores; use MixedCaps",
		}
		return &d
	}

	for _, acr := range commonAcronyms {
		lower := strings.ToLower(acr)
		mixed := strings.ToUpper(acr[:1]) + strings.ToLower(acr[1:])
		if strings.Contains(name, mixed) && !strings.Contains(name, acr) {
			idx := strings.Index(name, mixed)
			atEnd := idx+len(mixed) == len(name)
			nextIsUpper := !atEnd &&
				unicode.IsUpper(rune(name[idx+len(mixed)]))
			if atEnd || nextIsUpper {
				msg := "'" + mixed + "' in '" + name +
					"' should be '" + acr +
					"' (Go convention: " + lower + " -> " + acr + ")"
				d := rule.Diagnostic{
					Rule:     "naming-convention",
					Category: rule.CategoryStyle,
					Severity: rule.SeverityInfo,
					Pos:      ctx.FileSet.Position(ident.Pos()),
					End:      ctx.FileSet.Position(ident.End()),
					Message:  msg,
				}
				return &d
			}
		}
	}

	return nil
}

var commonAcronyms = []string{
	"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID",
	"HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS",
	"QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL",
	"SSH", "TCP", "TLS", "TTL", "UDP", "UI", "UID",
	"URI", "URL", "UTF8", "UUID", "VM", "XML", "XMPP",
	"XSRF", "XSS",
}

func init() {
	rule.Register(NamingConvention{})
}
