package astgrep

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// sarifSchema is the SARIF 2.1.0 schema URL.
const sarifSchema = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"

// sarifDocument is the SARIF 2.1.0 top-level document.
// https://docs.oasis-open.org/sarif/sarif/v2.1.0/
type sarifDocument struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

// sarifRun is the result of a single tool execution.
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

// sarifTool holds scan tool information.
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

// sarifDriver holds tool driver metadata.
type sarifDriver struct {
	Name    string       `json:"name"`
	Version string       `json:"version"`
	Rules   []sarifRule  `json:"rules,omitempty"`
}

// sarifRule is a SARIF rule definition.
type sarifRule struct {
	ID               string            `json:"id"`
	ShortDescription *sarifMessage     `json:"shortDescription,omitempty"`
	Properties       map[string]string `json:"properties,omitempty"`
}

// sarifResult is a single finding result.
type sarifResult struct {
	RuleID     string           `json:"ruleId"`
	Level      string           `json:"level"`
	Message    sarifMessage     `json:"message"`
	Locations  []sarifLocation  `json:"locations"`
	// Properties is a SARIF 2.1.0 §3.52 property bag.
	// Uses map[string]any to accommodate both existing string values and tags []string.
	// REQ-UTIL-002-005/006: when owasp/cwe keys are present, external/owasp/* and external/cwe/* tags are appended to tags.
	Properties map[string]any `json:"properties,omitempty"`
}

// sarifMessage is the SARIF message text.
type sarifMessage struct {
	Text string `json:"text"`
}

// sarifLocation holds code location information.
type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

// sarifPhysicalLocation holds file and region information.
type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

// sarifArtifactLocation is the file URI.
type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

// sarifRegion is the code region (line/column).
type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

// ToSARIF converts a slice of Findings to SARIF 2.1.0 JSON format.
// REQ-ASTG-UPG-022: SARIF 2.1.0 output support
//
// sgVersion is the ast-grep CLI version string (reflected in tool.driver.version).
// When findings is nil, a valid SARIF document with an empty results array is returned.
func ToSARIF(findings []Finding, sgVersion string) ([]byte, error) {
	if sgVersion == "" {
		sgVersion = "unknown"
	}

	// Collect rule list (dedup)
	ruleSet := make(map[string]sarifRule)
	for _, f := range findings {
		if f.RuleID == "" {
			continue
		}
		if _, exists := ruleSet[f.RuleID]; !exists {
			rule := sarifRule{
				ID: f.RuleID,
				ShortDescription: &sarifMessage{Text: f.Message},
			}
			if len(f.Metadata) > 0 {
				rule.Properties = f.Metadata
			}
			ruleSet[f.RuleID] = rule
		}
	}

	// @MX:NOTE: Sort ascending by rule ID to guarantee deterministic output.
	// Go map iteration order is non-deterministic, so without sorting the SARIF output
	// differs between runs: (1) snapshot tests become impossible, (2) GitHub Code Scanning
	// diffs become noisy, (3) CI reproducibility degrades. Sorting is mandatory. (issue #644)
	rules := make([]sarifRule, 0, len(ruleSet))
	for _, r := range ruleSet {
		rules = append(rules, r)
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID < rules[j].ID
	})

	// Convert to results
	results := make([]sarifResult, 0, len(findings))
	for _, f := range findings {
		result := sarifResult{
			RuleID:  f.RuleID,
			Level:   severityToSARIFLevel(f.Severity),
			Message: sarifMessage{Text: f.Message},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: toFileURI(f.File),
						},
						Region: sarifRegion{
							StartLine:   maxInt(f.Line, 1),
							StartColumn: f.Column,
							EndLine:     f.EndLine,
							EndColumn:   f.EndColumn,
						},
					},
				},
			},
		}

		// Pass metadata as SARIF properties and append OWASP/CWE tags.
		// REQ-UTIL-002-005/006: when owasp/cwe keys are present in Metadata, external/* tags are added.
		if len(f.Metadata) > 0 {
			props := make(map[string]any, len(f.Metadata)+1)
			for k, v := range f.Metadata {
				props[k] = v // preserve existing string values (backward compat)
			}
			// SARIF 2.1.0 §3.52: append external/owasp/* and external/cwe/* entries to tags array
			var tags []string
			if v, ok := f.Metadata["owasp"]; ok && v != "" {
				tags = append(tags, "external/owasp/"+sanitizeTagValue(v))
			}
			if v, ok := f.Metadata["cwe"]; ok && v != "" {
				tags = append(tags, "external/cwe/"+sanitizeTagValue(v))
			}
			if len(tags) > 0 {
				props["tags"] = tags
			}
			result.Properties = props
		}

		results = append(results, result)
	}

	doc := sarifDocument{
		Schema:  sarifSchema,
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "moai-ast-grep",
						Version: sgVersion,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	output, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("SARIF serialization: %w", err)
	}

	return output, nil
}

// severityToSARIFLevel converts an ast-grep severity to a SARIF level.
// Supported SARIF 2.1.0 levels: "error", "warning", "note", "none"
// REQ-ASTG-UPG-022: severity mapping (error→error, warning→warning, info→note)
func severityToSARIFLevel(severity string) string {
	switch strings.ToLower(severity) {
	case "error":
		return "error"
	case "warning", "warn":
		return "warning"
	default:
		return "note" // info, hint, and "" all map to note
	}
}

// toFileURI converts a file path to the SARIF URI format.
// Absolute paths are used as-is without adding the file:// scheme.
func toFileURI(path string) string {
	if path == "" {
		return ""
	}
	// Already in URI format — return as-is.
	if strings.HasPrefix(path, "file://") {
		return path
	}
	return path
}

// maxInt returns the larger of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// sanitizeTagValue converts an OWASP/CWE metadata value into a form safe
// for use as a SARIF tags entry. Transformation rules:
//  1. Convert to lowercase
//  2. Replace non-alphanumeric characters with hyphens
//  3. Collapse consecutive hyphens to a single hyphen
//  4. Strip leading and trailing hyphens
//
// Examples: "A03:2021 - Injection" → "a03-2021-injection", "CWE-89" → "cwe-89"
func sanitizeTagValue(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteByte('-')
		}
	}
	// Collapse consecutive hyphens
	result := b.String()
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}
