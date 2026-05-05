package spec

import "encoding/json"

// sarif.go generates output in SARIF 2.1.0 format.
// SPEC-V3R2-SPC-003 REQ-SPC-003-021 implementation.
//
// SARIF 2.1.0 schema: https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json

const (
	sarifVersion = "2.1.0"
	sarifSchema  = "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json"
	toolName     = "moai-spec-lint"
	toolVersion  = "0.1.0"
)

// sarifLog is the top-level struct for SARIF 2.1.0
type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

// sarifRun represents a single lint execution
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

// sarifTool is lint tool information
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

// sarifDriver is lint tool driver information
type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []sarifRule `json:"rules"`
}

// sarifRule is lint rule information
type sarifRule struct {
	ID               string           `json:"id"`
	ShortDescription sarifMessage     `json:"shortDescription"`
	DefaultConfig    sarifRuleDefault `json:"defaultConfiguration"`
}

// sarifRuleDefault is the default configuration for rule
type sarifRuleDefault struct {
	Level string `json:"level"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

// sarifMessage is SARIF message struct
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

// sarifArtifactLocation is file URI
type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

// severityToSARIFLevel converts Severity to SARIF level string
func severityToSARIFLevel(sev Severity) string {
	switch sev {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return "note"
	}
}

// marshalSARIF converts findings to JSON bytes in SARIF 2.1.0 format
func marshalSARIF(findings []Finding) ([]byte, error) {
	rulesSeen := make(map[string]bool)
	var rules []sarifRule
	for _, f := range findings {
		if !rulesSeen[f.Code] {
			rulesSeen[f.Code] = true
			rules = append(rules, sarifRule{
				ID: f.Code,
				ShortDescription: sarifMessage{
					Text: f.Code,
				},
				DefaultConfig: sarifRuleDefault{
					Level: severityToSARIFLevel(f.Severity),
				},
			})
		}
	}

	// Build results
	results := make([]sarifResult, 0, len(findings))
	for _, f := range findings {
		result := sarifResult{
			RuleID: f.Code,
			Level:  severityToSARIFLevel(f.Severity),
			Message: sarifMessage{
				Text: f.Message,
			},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: f.File,
						},
						Region: sarifRegion{
							StartLine: positiveLineNum(f.Line),
						},
					},
				},
			},
		}
		results = append(results, result)
	}

	log := sarifLog{
		Schema:  sarifSchema,
		Version: sarifVersion,
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    toolName,
						Version: toolVersion,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	return json.MarshalIndent(log, "", "  ")
}

func positiveLineNum(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
