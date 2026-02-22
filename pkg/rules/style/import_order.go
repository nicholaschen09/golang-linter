package style

import (
	"go/ast"
	"strings"

	"github.com/nicholas/glint/pkg/rule"
)

type ImportOrder struct{}

func (ImportOrder) Name() string            { return "import-order" }
func (ImportOrder) Category() rule.Category { return rule.CategoryStyle }
func (ImportOrder) Severity() rule.Severity { return rule.SeverityInfo }
func (ImportOrder) Description() string {
	return "Enforces import grouping: stdlib, then external, then internal"
}
func (ImportOrder) NeedsTypeInfo() bool { return false }
func (ImportOrder) NodeTypes() []ast.Node {
	return nil // uses FileRule interface
}

func (r ImportOrder) Check(_ *rule.Context, _ ast.Node) []rule.Diagnostic {
	return nil
}

func (ImportOrder) CheckFile(ctx *rule.Context) []rule.Diagnostic {
	var diags []rule.Diagnostic

	for _, decl := range ctx.File.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		var imports []importInfo
		for _, spec := range gd.Specs {
			is, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			path := strings.Trim(is.Path.Value, `"`)
			imports = append(imports, importInfo{
				path:  path,
				group: classifyImport(path),
				spec:  is,
			})
		}

		if len(imports) <= 1 {
			continue
		}

		// Verify ordering: stdlib (0) < external (1) < internal (2)
		lastGroup := -1
		for _, imp := range imports {
			if imp.group < lastGroup {
				diags = append(diags, rule.Diagnostic{
					Rule:     "import-order",
					Category: rule.CategoryStyle,
					Severity: rule.SeverityInfo,
					Pos:      ctx.FileSet.Position(imp.spec.Pos()),
					End:      ctx.FileSet.Position(imp.spec.End()),
					Message:  "import '" + imp.path + "' is out of order; expected grouping: stdlib, external, internal",
				})
			}
			if imp.group > lastGroup {
				lastGroup = imp.group
			}
		}
	}

	return diags
}

type importInfo struct {
	path  string
	group int // 0=stdlib, 1=external, 2=internal
	spec  *ast.ImportSpec
}

func classifyImport(path string) int {
	if !strings.Contains(path, ".") {
		return 0 // stdlib
	}
	return 1 // external (internal detection would need module path)
}

func init() {
	rule.Register(ImportOrder{})
}
