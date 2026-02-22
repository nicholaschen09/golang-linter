package report

import (
	"encoding/json"
	"io"

	"github.com/nicholas/glint/pkg/rule"
)

type SARIFReporter struct{}

type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []sarifRule `json:"rules,omitempty"`
}

type sarifRule struct {
	ID               string          `json:"id"`
	ShortDescription sarifMessage    `json:"shortDescription"`
}

type sarifResult struct {
	RuleID    string           `json:"ruleId"`
	Level     string           `json:"level"`
	Message   sarifMessage     `json:"message"`
	Locations []sarifLocation  `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}

func (r *SARIFReporter) Report(w io.Writer, diagnostics []rule.Diagnostic) error {
	ruleMap := make(map[string]bool)
	var rules []sarifRule
	results := make([]sarifResult, 0, len(diagnostics))

	for _, d := range diagnostics {
		if !ruleMap[d.Rule] {
			ruleMap[d.Rule] = true
			rules = append(rules, sarifRule{
				ID:               d.Rule,
				ShortDescription: sarifMessage{Text: d.Rule},
			})
		}

		results = append(results, sarifResult{
			RuleID:  d.Rule,
			Level:   sarifLevel(d.Severity),
			Message: sarifMessage{Text: d.Message},
			Locations: []sarifLocation{{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: d.Pos.Filename},
					Region: sarifRegion{
						StartLine:   d.Pos.Line,
						StartColumn: d.Pos.Column,
					},
				},
			}},
		})
	}

	log := sarifLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Runs: []sarifRun{{
			Tool: sarifTool{
				Driver: sarifDriver{
					Name:    "glint",
					Version: "0.1.0",
					Rules:   rules,
				},
			},
			Results: results,
		}},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

func sarifLevel(s rule.Severity) string {
	switch s {
	case rule.SeverityError:
		return "error"
	case rule.SeverityWarning:
		return "warning"
	default:
		return "note"
	}
}
