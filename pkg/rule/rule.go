package rule

import (
	"go/ast"
	"go/token"
	"go/types"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

type Category int

const (
	CategoryBugs Category = iota
	CategoryStyle
	CategoryPerf
	CategorySecurity
)

func (c Category) String() string {
	switch c {
	case CategoryBugs:
		return "bugs"
	case CategoryStyle:
		return "style"
	case CategoryPerf:
		return "perf"
	case CategorySecurity:
		return "security"
	default:
		return "unknown"
	}
}

type Diagnostic struct {
	Rule     string
	Category Category
	Severity Severity
	Pos      token.Position
	End      token.Position
	Message  string
}

type Context struct {
	File     *ast.File
	FileSet  *token.FileSet
	TypeInfo *types.Info
	Pkg      *types.Package
	FileHash string
	FilePath string
}

// Rule is the interface that all lint rules must implement.
type Rule interface {
	Name() string
	Category() Category
	Severity() Severity
	Description() string
	NeedsTypeInfo() bool
	// NodeTypes returns zero-value instances of the AST node types
	// this rule is interested in. The walker uses reflect.TypeOf on
	// each to build a dispatch table.
	NodeTypes() []ast.Node
	Check(ctx *Context, node ast.Node) []Diagnostic
}

// FileRule is an optional interface for rules that want to inspect
// the entire file at once rather than individual nodes.
type FileRule interface {
	Rule
	CheckFile(ctx *Context) []Diagnostic
}
