package astgrep_test

import (
	"encoding/json"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// TestToSARIF_EmptyFindings: verifies that a valid SARIF document is generated when there are no findings
func TestToSARIF_EmptyFindings(t *testing.T) {
	output, err := astgrep.ToSARIF(nil, "1.0.0-test")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}
	if len(output) == 0 {
		t.Fatal("ToSARIF() returned empty output")
	}

	var doc map[string]any
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("SARIF output is not valid JSON: %v", err)
	}

	// validate $schema field
	schema, ok := doc["$schema"].(string)
	if !ok || schema == "" {
		t.Error("SARIF $schema field is missing or empty")
	}

	// validate version field
	version, ok := doc["version"].(string)
	if !ok || version != "2.1.0" {
		t.Errorf("SARIF version = %q, want 2.1.0", version)
	}

	// validate runs array
	runs, ok := doc["runs"].([]any)
	if !ok || len(runs) == 0 {
		t.Fatal("SARIF runs array is missing or empty")
	}
}

// TestToSARIF_ToolDriver: verifies that tool.driver fields satisfy SPEC requirements (AC5)
func TestToSARIF_ToolDriver(t *testing.T) {
	output, err := astgrep.ToSARIF(nil, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	var doc map[string]any
	_ = json.Unmarshal(output, &doc)

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)
	tool := run["tool"].(map[string]any)
	driver := tool["driver"].(map[string]any)

	// tool.driver.name must be "moai-ast-grep" (AC5)
	if name := driver["name"].(string); name != "moai-ast-grep" {
		t.Errorf("tool.driver.name = %q, want moai-ast-grep", name)
	}

	// tool.driver.version must reflect the passed version (AC5)
	if ver := driver["version"].(string); ver != "0.42.1" {
		t.Errorf("tool.driver.version = %q, want 0.42.1", ver)
	}
}

// TestToSARIF_FindingMapping: verifies that Findings are correctly mapped to SARIF results (AC5)
func TestToSARIF_FindingMapping(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:   "go-no-raw-getenv",
			Severity: "warning",
			Message:  "환경변수를 직접 사용하지 마세요",
			File:     "internal/cli/main.go",
			Line:     10,
			Column:   5,
		},
		{
			RuleID:   "sec-sql-injection",
			Severity: "error",
			Message:  "SQL 인젝션 가능성",
			File:     "internal/db/query.go",
			Line:     42,
		},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	var doc map[string]any
	_ = json.Unmarshal(output, &doc)

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)
	results, ok := run["results"].([]any)
	if !ok {
		t.Fatal("SARIF results array is missing")
	}

	if len(results) != 2 {
		t.Fatalf("SARIF results len = %d, want 2", len(results))
	}
}

// TestToSARIF_SeverityMapping: verifies that severity is correctly mapped to SARIF level (AC5)
func TestToSARIF_SeverityMapping(t *testing.T) {
	tests := []struct {
		severity  string
		wantLevel string
	}{
		{"error", "error"},
		{"warning", "warning"},
		{"info", "note"},
		{"", "note"},  // empty string → note (SARIF default)
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			findings := []astgrep.Finding{
				{
					RuleID:   "test-rule",
					Severity: tt.severity,
					Message:  "테스트",
					File:     "test.go",
					Line:     1,
				},
			}

			output, err := astgrep.ToSARIF(findings, "1.0.0")
			if err != nil {
				t.Fatalf("ToSARIF() error = %v", err)
			}

			var doc map[string]any
			_ = json.Unmarshal(output, &doc)

			runs := doc["runs"].([]any)
			run := runs[0].(map[string]any)
			results := run["results"].([]any)
			result := results[0].(map[string]any)

			level, ok := result["level"].(string)
			if !ok {
				t.Fatalf("SARIF result.level field is missing")
			}
			if level != tt.wantLevel {
				t.Errorf("SARIF level = %q, want %q (severity=%q)", level, tt.wantLevel, tt.severity)
			}
		})
	}
}

// TestToSARIF_MetadataPreservation: verifies that Finding metadata (CWE/OWASP) is passed through to SARIF (AC5)
func TestToSARIF_MetadataPreservation(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:   "sec-sql-injection",
			Severity: "error",
			Message:  "SQL 인젝션",
			File:     "db.go",
			Line:     1,
			Metadata: map[string]string{
				"owasp": "A03:2021",
				"cwe":   "CWE-89",
			},
		},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	// parse as JSON and verify presence of properties field
	var doc map[string]any
	_ = json.Unmarshal(output, &doc)

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)
	result := results[0].(map[string]any)

	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("SARIF result.properties field is missing")
	}

	if cwe, ok := props["cwe"].(string); !ok || cwe == "" {
		t.Error("SARIF result.properties.cwe field is missing or empty")
	}
}

// TestToSARIF_RulesDeterministicOrder: verifies that tool.driver.rules array is sorted
// ascending by rule ID and produces identical output on every run.
// REGRESSION: prevents the issue (issue #644) where non-deterministic Go map iteration
// caused SARIF output to vary between runs. Ensures stability for snapshot tests and
// GitHub Code Scanning diffs.
func TestToSARIF_RulesDeterministicOrder(t *testing.T) {
	// intentionally provide input in near-reverse-alphabetical order so that
	// output depending on map iteration order would be unstable.
	findings := []astgrep.Finding{
		{RuleID: "zeta-rule", Severity: "warning", Message: "zeta", File: "a.go", Line: 1},
		{RuleID: "alpha-rule", Severity: "error", Message: "alpha", File: "b.go", Line: 2},
		{RuleID: "mike-rule", Severity: "warning", Message: "mike", File: "c.go", Line: 3},
		{RuleID: "bravo-rule", Severity: "error", Message: "bravo", File: "d.go", Line: 4},
		{RuleID: "yankee-rule", Severity: "warning", Message: "yankee", File: "e.go", Line: 5},
		{RuleID: "charlie-rule", Severity: "error", Message: "charlie", File: "f.go", Line: 6},
	}

	wantOrder := []string{
		"alpha-rule",
		"bravo-rule",
		"charlie-rule",
		"mike-rule",
		"yankee-rule",
		"zeta-rule",
	}

	// must produce identical order on every call (validates avoidance of non-deterministic map iteration).
	const iterations = 20
	var firstOutput []byte

	for i := 0; i < iterations; i++ {
		output, err := astgrep.ToSARIF(findings, "1.0.0-test")
		if err != nil {
			t.Fatalf("iteration %d: ToSARIF() error = %v", i, err)
		}

		if i == 0 {
			firstOutput = output
		} else if string(output) != string(firstOutput) {
			t.Fatalf("iteration %d: SARIF output differs between runs (non-deterministic)\nfirst:\n%s\ngot:\n%s",
				i, string(firstOutput), string(output))
		}

		var doc map[string]any
		if err := json.Unmarshal(output, &doc); err != nil {
			t.Fatalf("iteration %d: JSON parse failed: %v", i, err)
		}

		runs := doc["runs"].([]any)
		run := runs[0].(map[string]any)
		tool := run["tool"].(map[string]any)
		driver := tool["driver"].(map[string]any)
		rules, ok := driver["rules"].([]any)
		if !ok {
			t.Fatalf("iteration %d: tool.driver.rules array is missing", i)
		}

		if len(rules) != len(wantOrder) {
			t.Fatalf("iteration %d: rules len = %d, want %d", i, len(rules), len(wantOrder))
		}

		for idx, rule := range rules {
			r := rule.(map[string]any)
			id := r["id"].(string)
			if id != wantOrder[idx] {
				t.Errorf("iteration %d: rules[%d].id = %q, want %q (entire order must be ascending by ID)",
					i, idx, id, wantOrder[idx])
			}
		}
	}
}

// TestToSARIF_OWASPTag verifies that a Finding with metadata.owasp emits an
// external/owasp/* tag in properties.tags. AC-UTIL-002-05
func TestToSARIF_OWASPTag(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:   "sec-injection",
			Severity: "error",
			Message:  "injection risk",
			File:     "db.go",
			Line:     1,
			Metadata: map[string]string{
				"owasp": "A03:2021 - Injection",
			},
		},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	var doc map[string]any
	_ = json.Unmarshal(output, &doc)

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)
	result := results[0].(map[string]any)

	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("SARIF result.properties field is missing")
	}

	tags, ok := props["tags"].([]any)
	if !ok {
		t.Fatal("SARIF result.properties.tags field is missing or not an array")
	}

	found := false
	for _, tag := range tags {
		s, ok := tag.(string)
		if !ok {
			continue
		}
		if len(s) > len("external/owasp/") && s[:len("external/owasp/")] == "external/owasp/" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected tag starting with external/owasp/ in properties.tags, got: %v", tags)
	}
}

// TestToSARIF_CWETag verifies that a Finding with metadata.cwe emits an
// external/cwe/* tag in properties.tags. AC-UTIL-002-06
func TestToSARIF_CWETag(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:   "sec-sqli",
			Severity: "error",
			Message:  "SQL injection",
			File:     "db.go",
			Line:     1,
			Metadata: map[string]string{
				"cwe": "CWE-89",
			},
		},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	var doc map[string]any
	_ = json.Unmarshal(output, &doc)

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)
	result := results[0].(map[string]any)

	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("SARIF result.properties field is missing")
	}

	tags, ok := props["tags"].([]any)
	if !ok {
		t.Fatal("SARIF result.properties.tags field is missing or not an array")
	}

	found := false
	for _, tag := range tags {
		s, ok := tag.(string)
		if !ok {
			continue
		}
		if s == "external/cwe/cwe-89" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected external/cwe/cwe-89 in properties.tags, got: %v", tags)
	}
}

// TestSARIF_BackwardCompatibility locks existing SARIF field names and types.
// Any rename or removal of existing fields MUST fail this test.
// Cross-cutting invariant for REQ-UTIL-002-021.
func TestSARIF_BackwardCompatibility(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:    "bc-rule",
			Severity:  "error",
			Message:   "backward compat message",
			File:      "test.go",
			Line:      10,
			Column:    5,
			EndLine:   10,
			EndColumn: 20,
			Metadata: map[string]string{
				"owasp": "A03:2021",
				"cwe":   "CWE-89",
			},
		},
	}

	output, err := astgrep.ToSARIF(findings, "1.0.0-bc-test")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Top-level fields
	if doc["$schema"] == nil {
		t.Error("BACKWARD COMPAT: $schema field removed or renamed")
	}
	if doc["version"] == nil {
		t.Error("BACKWARD COMPAT: version field removed or renamed")
	}
	if doc["runs"] == nil {
		t.Error("BACKWARD COMPAT: runs field removed or renamed")
	}

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)

	// tool.driver fields
	tool := run["tool"].(map[string]any)
	driver := tool["driver"].(map[string]any)
	if driver["name"] == nil {
		t.Error("BACKWARD COMPAT: tool.driver.name removed")
	}
	if driver["version"] == nil {
		t.Error("BACKWARD COMPAT: tool.driver.version removed")
	}

	// result fields
	results := run["results"].([]any)
	result := results[0].(map[string]any)

	if result["ruleId"] == nil {
		t.Error("BACKWARD COMPAT: result.ruleId removed")
	}
	if result["level"] == nil {
		t.Error("BACKWARD COMPAT: result.level removed")
	}
	if result["message"] == nil {
		t.Error("BACKWARD COMPAT: result.message removed")
	}
	if result["locations"] == nil {
		t.Error("BACKWARD COMPAT: result.locations removed")
	}

	// properties: existing metadata keys MUST still be present as top-level string values
	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("BACKWARD COMPAT: result.properties field removed or type changed")
	}
	if _, ok := props["owasp"].(string); !ok {
		t.Error("BACKWARD COMPAT: result.properties.owasp removed or type changed from string")
	}
	if _, ok := props["cwe"].(string); !ok {
		t.Error("BACKWARD COMPAT: result.properties.cwe removed or type changed from string")
	}

	// NEW: properties.tags array must be present (additive)
	tags, ok := props["tags"].([]any)
	if !ok {
		t.Error("NEW FIELD: result.properties.tags missing or not an array")
	}
	if len(tags) == 0 {
		t.Error("NEW FIELD: result.properties.tags should contain OWASP and CWE entries")
	}
}

// TestToSARIF_RoundTrip: verifies that ToSARIF output conforms to valid SARIF 2.1.0 schema
func TestToSARIF_RoundTrip(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:   "test-rule-1",
			Severity: "error",
			Message:  "첫 번째 테스트 규칙",
			File:     "file1.go",
			Line:     10,
			Column:   5,
			EndLine:  10,
			EndColumn: 20,
		},
		{
			RuleID:   "test-rule-2",
			Severity: "warning",
			Message:  "두 번째 테스트 규칙",
			File:     "file2.go",
			Line:     42,
		},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	// validate round-trip via JSON parse and re-serialization
	var doc any
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("ToSARIF() output is not valid JSON: %v", err)
	}

	reEncoded, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("re-serialization failed: %v", err)
	}

	var redoc map[string]any
	if err := json.Unmarshal(reEncoded, &redoc); err != nil {
		t.Fatalf("re-parsing failed: %v", err)
	}

	// re-verify core fields after round-trip
	if redoc["version"] != "2.1.0" {
		t.Errorf("version after round-trip = %v, want 2.1.0", redoc["version"])
	}
}
