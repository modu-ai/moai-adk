package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/constitution"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// assertExitCode fails the test if err does not carry the expected exit code via
// the ExitCoder boundary (cmd/moai/main.go errors.As mapping).
func assertExitCode(t *testing.T, err error, want int) {
	t.Helper()
	var ec exitCoder
	if !errors.As(err, &ec) {
		t.Fatalf("error does not satisfy ExitCoder: got %T (%v); want ExitCode()==%d", err, err, want)
	}
	if got := ec.ExitCode(); got != want {
		t.Fatalf("ExitCode() = %d, want %d (err=%v)", got, want, err)
	}
}

// TestExitCodeContract_SpecLintInvalidArgs verifies REQ-CONT-001-005: moai spec
// lint with mutually-exclusive --json and --sarif exits 3 (invalid arguments).
func TestExitCodeContract_SpecLintInvalidArgs(t *testing.T) {
	cmd := newSpecLintCmd()
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("set --json: %v", err)
	}
	if err := cmd.Flags().Set("sarif", "true"); err != nil {
		t.Fatalf("set --sarif: %v", err)
	}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{})
	assertExitCode(t, err, 3)
}

// TestExitCodeContract_SpecAuditMustFix verifies REQ-CONT-001-005: moai spec
// audit --strict with a MUST-FIX drift finding exits 2 (system-error class).
func TestExitCodeContract_SpecAuditMustFix(t *testing.T) {
	result := &spec.AuditResult{
		DriftFindings: []spec.DriftFinding{
			{SpecID: "SPEC-TEST-001", Severity: "MUST-FIX", FindingType: "StatusDrift"},
		},
	}
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	// strict=true + a MUST-FIX finding → exit 2.
	err := renderAuditResult(cmd, result, false, true)
	assertExitCode(t, err, 2)

	// No MUST-FIX finding → nil (exit 0).
	cleanResult := &spec.AuditResult{
		DriftFindings: []spec.DriftFinding{
			{SpecID: "SPEC-TEST-002", Severity: "INFO"},
		},
	}
	if err := renderAuditResult(cmd, cleanResult, false, true); err != nil {
		t.Fatalf("renderAuditResult clean result returned error: %v", err)
	}
}

// TestExitCodeContract_Constitution verifies REQ-CONT-001-005: constitution
// validate against a registry entry whose source file is missing exits 2
// (SOURCE_FILE_MISSING → exitCodeError{2}, mapped via the ExitCoder boundary
// after the M1 ExitCode() method addition).
func TestExitCodeContract_Constitution(t *testing.T) {
	dir := t.TempDir()
	// Registry entry references nonexistent.md → SOURCE_FILE_MISSING. The
	// registry loader extracts YAML from a ```yaml code fence in a .md file.
	regBody := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: nonexistent.md
  anchor: "#rules"
  clause: "Some rule."
  canary_gate: true
`
	regContent := "# Registry\n\n```yaml\n" + regBody + "\n```\n"
	regPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(regPath, []byte(regContent), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	var out, errBuf bytes.Buffer
	err := runConstitutionValidate(&out, &errBuf, dir, regPath, constitution.ValidateOptions{}, "text")
	assertExitCode(t, err, 2)
}

// TestAstgrepExitCode verifies REQ-CONT-001-004: an ast-grep scan with
// error-severity findings exits 1 under text, json, AND sarif formats alike
// (HasErrors evaluated independently of the output-format branch).
func TestAstgrepExitCode(t *testing.T) {
	for _, format := range []string{"text", "json", "sarif"} {
		t.Run(format, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Rule with severity: error that matches a magic string literal. (sg
			// in this environment reliably matches string-literal patterns;
			// call-expression patterns like fmt.Println($X) do not match.)
			rulesDir := filepath.Join(tmpDir, "rules", "go")
			if err := os.MkdirAll(rulesDir, 0o755); err != nil {
				t.Fatalf("mkdir rules: %v", err)
			}
			ruleYAML := "---\nid: no-magic-token\nlanguage: go\nseverity: error\n" +
				"message: \"magic token forbidden\"\npattern: '\"DIAG_TOKEN\"'\n"
			if err := os.WriteFile(filepath.Join(rulesDir, "test.yml"), []byte(ruleYAML), 0o644); err != nil {
				t.Fatalf("write rule: %v", err)
			}

			// Source file that triggers the rule.
			srcDir := filepath.Join(tmpDir, "src")
			if err := os.MkdirAll(srcDir, 0o755); err != nil {
				t.Fatalf("mkdir src: %v", err)
			}
			if err := os.WriteFile(filepath.Join(srcDir, "main.go"),
				[]byte("package main\n\nfunc main() { _ = \"DIAG_TOKEN\" }\n"), 0o644); err != nil {
				t.Fatalf("write src: %v", err)
			}

			var buf bytes.Buffer
			cmd := NewAstGrepCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{srcDir})
			if err := cmd.Flags().Set("format", format); err != nil {
				t.Fatalf("set format: %v", err)
			}
			if err := cmd.Flags().Set("rules-dir", filepath.Join(tmpDir, "rules")); err != nil {
				t.Fatalf("set rules-dir: %v", err)
			}

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected exit-1 ExitCoder for %s format with error findings, got nil", format)
			}
			assertExitCode(t, err, 1)
		})
	}
}

// TestHelpExitContract verifies REQ-CONT-001-008: for each changed command, the
// exit code DECLARED in the help text matches the exit code PRODUCED. This ties
// the documentation to the implementation so declaration drift is caught.
func TestHelpExitContract(t *testing.T) {
	t.Run("SpecLint", func(t *testing.T) {
		cmd := newSpecLintCmd()
		// Help declares "3 = invalid arguments".
		if !strings.Contains(cmd.Long, "3 = invalid arguments") {
			t.Fatalf("spec lint help must declare '3 = invalid arguments'; Long=\n%s", cmd.Long)
		}
		// Produced: --json + --sarif → exit 3.
		lint := newSpecLintCmd()
		_ = lint.Flags().Set("json", "true")
		_ = lint.Flags().Set("sarif", "true")
		lint.SetOut(&bytes.Buffer{})
		lint.SetErr(&bytes.Buffer{})
		assertExitCode(t, lint.RunE(lint, []string{}), 3)
	})

	t.Run("SpecAudit", func(t *testing.T) {
		cmd := newSpecAuditCmd()
		// Help declares "2 = strict mode + MUST-FIX".
		if !strings.Contains(cmd.Long, "2 = strict mode + MUST-FIX") {
			t.Fatalf("spec audit help must declare '2 = strict mode + MUST-FIX'; Long=\n%s", cmd.Long)
		}
		// Produced: MUST-FIX finding + strict → exit 2.
		result := &spec.AuditResult{
			DriftFindings: []spec.DriftFinding{
				{SpecID: "SPEC-TEST-001", Severity: "MUST-FIX"},
			},
		}
		c := &cobra.Command{}
		c.SetOut(&bytes.Buffer{})
		c.SetErr(&bytes.Buffer{})
		assertExitCode(t, renderAuditResult(c, result, false, true), 2)
	})

	t.Run("Constitution", func(t *testing.T) {
		cmd := newConstitutionValidateCmd()
		// Help declares "2=fatal (missing source file)".
		if !strings.Contains(cmd.Long, "2=fatal") && !strings.Contains(cmd.Long, "2 = fatal") {
			t.Fatalf("constitution validate help must declare exit 2 for missing source; Long=\n%s", cmd.Long)
		}
		// Produced: missing source file → exit 2.
		dir := t.TempDir()
		regBody := "- id: CONST-V3R2-001\n  zone: Frozen\n  zone_class: frozen-canonical\n  file: nonexistent.md\n  anchor: \"#r\"\n  clause: \"r.\"\n  canary_gate: true\n"
		regContent := "# Registry\n\n```yaml\n" + regBody + "\n```\n"
		regPath := filepath.Join(dir, "zone-registry.md")
		if err := os.WriteFile(regPath, []byte(regContent), 0o644); err != nil {
			t.Fatalf("write registry: %v", err)
		}
		var out, errBuf bytes.Buffer
		assertExitCode(t, runConstitutionValidate(&out, &errBuf, dir, regPath, constitution.ValidateOptions{}, "text"), 2)
	})
}
