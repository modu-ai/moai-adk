package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli"
)

// requireSG skips the test when the ast-grep binary is unavailable. The rewrite
// path shells out to `sg`, so without it there is nothing to assert.
func requireSG(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg binary not installed; skipping rewrite-path test")
	}
}

// TestAstEditCmd_Registered verifies the command exists and is named ast-edit.
// REQ-AGE-001 / AC-010.
func TestAstEditCmd_Registered(t *testing.T) {
	cmd := cli.NewAstEditCmd()
	if cmd == nil {
		t.Fatal("NewAstEditCmd() returned nil")
	}
	if !strings.Contains(cmd.Use, "ast-edit") {
		t.Errorf("AstEditCmd.Use = %q; must contain 'ast-edit'", cmd.Use)
	}
}

// TestAstEditCmd_HelpStatesWriteNature verifies the help text tells the user that
// a non-dry run modifies files in place. REQ-AGE-002 / AC-010.
func TestAstEditCmd_HelpStatesWriteNature(t *testing.T) {
	cmd := cli.NewAstEditCmd()
	long := strings.ToLower(cmd.Long)

	if !strings.Contains(long, "modif") {
		t.Errorf("AstEditCmd.Long must state that files are modified; got %q", cmd.Long)
	}
	if !strings.Contains(long, "--dry") {
		t.Errorf("AstEditCmd.Long must mention --dry as the preview path; got %q", cmd.Long)
	}
}

// TestAstEditCmd_Flags verifies the flag surface. REQ-AGE-002/003/004.
func TestAstEditCmd_Flags(t *testing.T) {
	cmd := cli.NewAstEditCmd()

	for _, flag := range []string{"dry", "pattern", "rewrite", "lang", "rules-dir", "rule", "format"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag --%s is not registered", flag)
		}
	}
}

// TestAstEditCmd_PatternWithoutRewriteRejected verifies that --pattern requires
// --rewrite. REQ-AGE-003 / AC-021.
func TestAstEditCmd_PatternWithoutRewriteRejected(t *testing.T) {
	cmd := cli.NewAstEditCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--pattern", "foo($A)", "."})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error when --pattern is given without --rewrite")
	}
	if !strings.Contains(err.Error(), "--rewrite") {
		t.Errorf("error must name the missing --rewrite flag; got %q", err.Error())
	}
}

// TestAstEditCmd_DryRunLeavesFileUnchanged verifies that a dry run reports matches
// without writing. REQ-AGE-002 / AC-012.
func TestAstEditCmd_DryRunLeavesFileUnchanged(t *testing.T) {
	requireSG(t)

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "sample.go")
	original := "package sample\n\nfunc run() {\n\tprintln(\"before\")\n}\n"
	if err := os.WriteFile(target, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	cmd := cli.NewAstEditCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"--dry",
		"--pattern", `println("before")`,
		"--rewrite", `println("after")`,
		"--lang", "go",
		target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("dry run returned an error: %v", err)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read fixture back: %v", err)
	}
	if string(after) != original {
		t.Errorf("dry run modified the file\n--- before ---\n%s\n--- after ---\n%s", original, string(after))
	}
}

// writeRuleFixture writes a flat-form rule file (the shape the loader reads) and
// returns the rules directory.
func writeRuleFixture(t *testing.T, body string) string {
	t.Helper()
	rulesDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(rulesDir, "fixture.yml"), []byte(body), 0644); err != nil {
		t.Fatalf("failed to write rule fixture: %v", err)
	}
	return rulesDir
}

// TestAstEditCmd_RuleModeAppliesFix verifies that a rule carrying a fix: is applied
// and that a detection-only rule is skipped with a notice rather than erroring.
// REQ-AGE-004 / AC-030.
func TestAstEditCmd_RuleModeAppliesFix(t *testing.T) {
	requireSG(t)

	rulesDir := writeRuleFixture(t, `id: fixture-rewrites
language: go
severity: warning
message: "rewritable"
pattern: println("before")
fix: println("after")
---
id: fixture-detection-only
language: go
severity: warning
message: "no fix field"
pattern: panic("boom")
`)

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "sample.go")
	original := "package sample\n\nfunc run() {\n\tprintln(\"before\")\n}\n"
	if err := os.WriteFile(target, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	var out bytes.Buffer
	cmd := cli.NewAstEditCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--rules-dir", rulesDir, "--lang", "go", target})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("rule-mode run returned an error: %v", err)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read fixture back: %v", err)
	}
	if !strings.Contains(string(after), `println("after")`) {
		t.Errorf("rule fix was not applied:\n%s", string(after))
	}
	if !strings.Contains(out.String(), "detection-only") {
		t.Errorf("expected a skip notice for the detection-only rule; got %q", out.String())
	}
}

// TestAstEditCmd_UnreadableMatcherReportedSeparately verifies that a rule whose
// matcher the loader cannot read — ast-grep's nested `rule:` block, which leaves
// Pattern empty because only the flat top-level `pattern:` field is parsed — is
// reported as a loader limitation and NOT misattributed to a missing fix:.
// Every shipped rule uses the nested form, so this is the path a user actually
// hits when running rule mode against the shipped ruleset. REQ-AGE-004.
func TestAstEditCmd_UnreadableMatcherReportedSeparately(t *testing.T) {
	requireSG(t)

	// The rule declares a fix:, so any notice blaming a missing fix: is wrong.
	rulesDir := writeRuleFixture(t, `id: fixture-nested-matcher
language: go
severity: warning
message: "nested rule block"
rule:
  pattern: panic("boom")
fix: panic("handled")
`)

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "sample.go")
	original := "package sample\n\nfunc run() {\n\tpanic(\"boom\")\n}\n"
	if err := os.WriteFile(target, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	var out bytes.Buffer
	cmd := cli.NewAstEditCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--rules-dir", rulesDir, "--lang", "go", target})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("rule-mode run returned an error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "matcher could not be read") {
		t.Errorf("expected the unreadable-matcher notice; got %q", got)
	}
	if strings.Contains(got, "no fix: field") {
		t.Errorf("a rule that declares fix: was reported as missing one; got %q", got)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read fixture back: %v", err)
	}
	if string(after) != original {
		t.Errorf("a rule with an unreadable matcher must not be applied:\n%s", string(after))
	}
}

// TestAstEditCmd_RuleFilterNarrowsToOneRule verifies that --rule restricts the pass.
// REQ-AGE-004 / AC-031.
func TestAstEditCmd_RuleFilterNarrowsToOneRule(t *testing.T) {
	requireSG(t)

	rulesDir := writeRuleFixture(t, `id: fixture-alpha
language: go
severity: warning
message: "alpha"
pattern: println("alpha")
fix: println("ALPHA")
---
id: fixture-beta
language: go
severity: warning
message: "beta"
pattern: println("beta")
fix: println("BETA")
`)

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "sample.go")
	original := "package sample\n\nfunc run() {\n\tprintln(\"alpha\")\n\tprintln(\"beta\")\n}\n"
	if err := os.WriteFile(target, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	cmd := cli.NewAstEditCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--rules-dir", rulesDir, "--rule", "fixture-alpha", "--lang", "go", target})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("rule-filter run returned an error: %v", err)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read fixture back: %v", err)
	}
	if !strings.Contains(string(after), `println("ALPHA")`) {
		t.Errorf("the selected rule was not applied:\n%s", string(after))
	}
	if !strings.Contains(string(after), `println("beta")`) {
		t.Errorf("--rule did not narrow the pass; the unselected rule was applied too:\n%s", string(after))
	}
}

// TestAstEditCmd_PatternModeRewritesFile verifies that a non-dry pattern-mode run
// applies the rewrite. REQ-AGE-003 / AC-020.
func TestAstEditCmd_PatternModeRewritesFile(t *testing.T) {
	requireSG(t)

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "sample.go")
	original := "package sample\n\nfunc run() {\n\tprintln(\"before\")\n}\n"
	if err := os.WriteFile(target, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	cmd := cli.NewAstEditCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"--pattern", `println("before")`,
		"--rewrite", `println("after")`,
		"--lang", "go",
		target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("pattern-mode run returned an error: %v", err)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read fixture back: %v", err)
	}
	if strings.Contains(string(after), `println("before")`) {
		t.Errorf("rewrite did not apply; file still contains the original call:\n%s", string(after))
	}
	if !strings.Contains(string(after), `println("after")`) {
		t.Errorf("rewrite did not apply; file lacks the replacement call:\n%s", string(after))
	}
}
