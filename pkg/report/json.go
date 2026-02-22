package report

import (
	"encoding/json"
	"io"

	"github.com/nicholas/glint/pkg/rule"
)

type JSONReporter struct{}

type jsonDiagnostic struct {
	Rule     string `json:"rule"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
}

func (r *JSONReporter) Report(w io.Writer, diagnostics []rule.Diagnostic) error {
	out := make([]jsonDiagnostic, 0, len(diagnostics))
	for _, d := range diagnostics {
		out = append(out, jsonDiagnostic{
			Rule:     d.Rule,
			Category: d.Category.String(),
			Severity: d.Severity.String(),
			File:     d.Pos.Filename,
			Line:     d.Pos.Line,
			Column:   d.Pos.Column,
			Message:  d.Message,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
