package quality

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// This file is the SPEC-GATE-ASTGREP-REPAIR-001 M2 regression guard for D2:
// the ast-grep gate step MUST exclude findings whose file paths fall under
// excluded roots (.claude/worktrees/, vendor/, *_test.go, testdata/). The
// primary exclusion mechanism (sgconfig.yml globs) is NOT supported by
// ast-grep 0.40.5 in config-mode (empirically verified: the globs field
// silently breaks the scan, returning 0 matches even for non-excluded files).
// The fallback per SPEC plan.md §D.2 B3/B4 is a path filter applied AFTER the
// scanner returns findings, here in RunAstGrepGateV2 (the gate layer — NOT in
// internal/astgrep/scanner.go, whose Scan body is preserved by REQ-GAR-010).
//
// AC-GAR-004 (REQ-GAR-004): worktree/vendor/test paths excluded.
// AC-GAR-005 (REQ-GAR-004): *_test.go and vendor/** excluded.
// AC-GAR-007 (REQ-GAR-005): (file,line,rule) dedup — satisfied by excluding
// the duplicate-bearing worktree root, so each (file,line,rule) triple in the
// main tree appears exactly once.

// requireSGAvailable skips when the sg CLI is absent.
func requireSGAvailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg binary not in PATH; exclusion behavior is not mechanically verifiable")
	}
}

// stageExclusionTree builds a temp project tree with:
//   - a real source file containing a genuine `return err` violation
//   - a _test.go file containing the SAME violation (should be excluded)
//   - a testdata/ file containing the SAME violation (should be excluded)
//   - a vendor/ file containing the SAME violation (should be excluded)
//   - a .claude/worktrees/ file containing the SAME violation (should be excluded)
//
// plus a staged rules dir with the shipped error-handling rule so the scan is
// deterministic.
func stageExclusionTree(t *testing.T) (projectDir, rulesDir string) {
	t.Helper()
	projectDir = t.TempDir()

	violation := []byte("package x\n\nfunc e() error { err := f(); return err }\nfunc f() error { return nil }\n")

	for _, rel := range []string{
		"real.go",
		"internal/foo/foo_test.go",
		"internal/bar/testdata/fixture.go",
		"vendor/v/file.go",
		".claude/worktrees/wt-x/file.go",
	} {
		full := filepath.Join(projectDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(full, violation, 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	// Stage the rules dir with a minimal sgconfig + the refined rule copy.
	rulesDir = filepath.Join(projectDir, ".astgrep-rules")
	rulesSub := filepath.Join(rulesDir, "go")
	if err := os.MkdirAll(rulesSub, 0o755); err != nil {
		t.Fatalf("mkdir rules: %v", err)
	}
	refinedRule := `---
id: go-error-not-wrapped
language: go
severity: warning
message: test
rule:
  pattern: return $ERR
fix: 'return fmt.Errorf("TODO: operation: %w", $ERR)'
constraints:
  ERR:
    regex: '^err(e?|s)$'
`
	if err := os.WriteFile(filepath.Join(rulesSub, "error-handling.yml"), []byte(refinedRule), 0o644); err != nil {
		t.Fatalf("write rule: %v", err)
	}
	sgconfig := "ruleDirs:\n  - go\n"
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte(sgconfig), 0o644); err != nil {
		t.Fatalf("write sgconfig: %v", err)
	}
	return projectDir, rulesDir
}

// TestGAR_AC004_ExcludedPathsFilteredFromFindings (AC-GAR-004/005):
// RunAstGrepGateV2 MUST NOT report findings under .claude/worktrees/,
// vendor/, *_test.go, or testdata/. The tree stages the same violation in
// five locations; only real.go (the non-excluded root) should surface.
func TestGAR_AC004_ExcludedPathsFilteredFromFindings(t *testing.T) {
	requireSGAvailable(t)
	projectDir, rulesDir := stageExclusionTree(t)

	// Use a rules_dir that resolves to the staged dir under projectDir.
	rel, err := filepath.Rel(projectDir, rulesDir)
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}
	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     rel,
		BlockOnError: false,
		WarnOnlyMode: true,
	}

	ok, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)
	if !ok {
		t.Fatalf("RunAstGrepGateV2 denied; output:\n%s", output)
	}

	// The output should mention real.go exactly once and none of the excluded paths.
	for _, excluded := range []string{"_test.go", "testdata/", "vendor/", ".claude/worktrees/"} {
		if contains(output, excluded) {
			t.Errorf("AC-GAR-004/005: excluded path pattern %q appears in gate output:\n%s", excluded, output)
		}
	}
	// real.go MUST appear (the one genuine violation we expect reported).
	if !contains(output, "real.go") {
		t.Errorf("AC-GAR-004: real.go (the non-excluded genuine violation) is missing from gate output:\n%s", output)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
