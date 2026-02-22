package report

import (
	"io"

	"github.com/nicholas/glint/pkg/rule"
)

type Reporter interface {
	Report(w io.Writer, diagnostics []rule.Diagnostic) error
}

func New(format string, color bool) Reporter {
	switch format {
	case "json":
		return &JSONReporter{}
	case "sarif":
		return &SARIFReporter{}
	default:
		return &TextReporter{Color: color}
	}
}
