package style

import (
	"bufio"
	"go/ast"
	"os"
	"strconv"

	"github.com/nicholas/glint/pkg/rule"
)

const defaultMaxLineLength = 120

type LineLength struct{}

func (LineLength) Name() string            { return "line-length" }
func (LineLength) Category() rule.Category { return rule.CategoryStyle }
func (LineLength) Severity() rule.Severity { return rule.SeverityWarning }
func (LineLength) Description() string {
	return "Reports lines exceeding a configurable maximum length"
}
func (LineLength) NeedsTypeInfo() bool  { return false }
func (LineLength) NodeTypes() []ast.Node { return nil }

func (LineLength) Check(_ *rule.Context, _ ast.Node) []rule.Diagnostic {
	return nil
}

func (LineLength) CheckFile(ctx *rule.Context) []rule.Diagnostic {
	f, err := os.Open(ctx.FilePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	maxLen := defaultMaxLineLength

	var diags []rule.Diagnostic
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if len(line) > maxLen {
			diags = append(diags, rule.Diagnostic{
				Rule:     "line-length",
				Category: rule.CategoryStyle,
				Severity: rule.SeverityWarning,
				Pos:      ctx.FileSet.Position(ctx.File.Pos()),
				Message:  ctx.FilePath + ":" + strconv.Itoa(lineNum) + " line is " + strconv.Itoa(len(line)) + " characters (max " + strconv.Itoa(maxLen) + ")",
			})
		}
	}

	return diags
}

func init() {
	rule.Register(LineLength{})
}
