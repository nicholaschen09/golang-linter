package report

import (
	"fmt"
	"io"

	"github.com/nicholas/glint/pkg/rule"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

type TextReporter struct {
	Color bool
}

func (r *TextReporter) Report(w io.Writer, diagnostics []rule.Diagnostic) error {
	for _, d := range diagnostics {
		if r.Color {
			sev := colorSeverity(d.Severity)
			_, _ = fmt.Fprintf(w, "%s%s%s: %s%s%s [%s%s%s] %s\n",
				colorGray, d.Pos, colorReset,
				sev, d.Severity, colorReset,
				colorCyan, d.Rule, colorReset,
				d.Message,
			)
		} else {
			_, _ = fmt.Fprintf(w, "%s: %s [%s] %s\n",
				d.Pos, d.Severity, d.Rule, d.Message,
			)
		}
	}

	if len(diagnostics) > 0 {
		_, _ = fmt.Fprintf(w, "\n%d issue(s) found.\n", len(diagnostics))
	}

	return nil
}

func colorSeverity(s rule.Severity) string {
	switch s {
	case rule.SeverityError:
		return colorRed
	case rule.SeverityWarning:
		return colorYellow
	case rule.SeverityInfo:
		return colorCyan
	default:
		return colorReset
	}
}
