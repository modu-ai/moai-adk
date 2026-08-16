package astgrep_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// This file is the SPEC-GATE-ASTGREP-REPAIR-001 regression guard for the
// refined `go-error-not-wrapped` ast-grep rule. It exercises the REAL shipped
// rule file (.moai/astgrep-rules/go/error-handling.yml) via the same
// `sg scan --config` path the quality gate uses at commit time, and asserts
// the three D1 acceptance criteria (AC-GAR-001/002/003).
//
// The fixtures under testdata/gar/ are intentionally minimal so the assertion
// is a binary PASS/FAIL rather than a noisy count over the live tree.

// requireSG skips the test when the ast-grep CLI is unavailable. The rule
// refinement is only mechanically verifiable when sg can actually run.
func requireSG(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg binary not in PATH; rule behavior is not mechanically verifiable")
	}
}

// projectRoot returns the repository root (the directory containing go.mod).
// The testdata fixtures live under the astgrep package directory; the rule
// file lives at <repo-root>/.moai/astgrep-rules/go/error-handling.yml — it
// moved out of .moai/config/ in #1557, because `moai update` wipes
// .moai/config/ wholesale and was deleting the local-only rules on every run.
func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	dir := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate go.mod above %s", wd)
	return ""
}

// realRuleFile returns the absolute path to the shipped error-handling rule
// file. The test copies it into a temp rules dir (alongside a minimal
// sgconfig.yml) so the Scanner exercises the exact YAML users ship, without
// depending on the wider astgrep-rules tree layout.
func realRuleFile(t *testing.T) string {
	t.Helper()
	root := projectRoot(t)
	p := filepath.Join(root, ".moai", "astgrep-rules", "go", "error-handling.yml")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("shipped rule file not found: %s", p)
	}
	return p
}

// stageRulesDir builds a temp rules directory containing a copy of the shipped
// error-handling.yml plus a minimal sgconfig.yml that points ruleDirs at the
// copy. The returned dir is a complete, self-contained config root for
// `sg scan --config <dir>/sgconfig.yml`.
func stageRulesDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	rulesSub := filepath.Join(tmp, "rules")
	if err := os.MkdirAll(rulesSub, 0o755); err != nil {
		t.Fatalf("mkdir rules: %v", err)
	}
	src, err := os.ReadFile(realRuleFile(t))
	if err != nil {
		t.Fatalf("read rule: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesSub, "rule.yml"), src, 0o644); err != nil {
		t.Fatalf("write rule copy: %v", err)
	}
	sgconfig := []byte("ruleDirs:\n  - rules\n")
	if err := os.WriteFile(filepath.Join(tmp, "sgconfig.yml"), sgconfig, 0o644); err != nil {
		t.Fatalf("write sgconfig: %v", err)
	}
	return tmp
}

// findingsFor scans the given fixture file through the staged config and
// returns the findings, filtered to the go-error-not-wrapped rule.
func findingsFor(t *testing.T, rulesDir, fixturePath string) []astgrep.Finding {
	t.Helper()
	cfg := &astgrep.ScannerConfig{
		RulesDir:     rulesDir,
		SGBinary:     "sg",
		WarnOnlyMode: true,
		Timeout:      10 * time.Second,
	}
	scanner := astgrep.NewScanner(cfg)
	all, err := scanner.Scan(context.Background(), fixturePath)
	if err != nil {
		t.Fatalf("Scan(%s): %v", fixturePath, err)
	}
	var out []astgrep.Finding
	for _, f := range all {
		if f.RuleID == "go-error-not-wrapped" {
			out = append(out, f)
		}
	}
	return out
}

// TestGAR_AC001_NegativeReturnsNotMatched (AC-GAR-001 / REQ-GAR-001):
// the shipped `go-error-not-wrapped` rule MUST NOT match non-error single
// literal/call returns. Before M1 the over-broad `return $ERR` pattern
// matched every one of these; this test fails RED until the rule is refined.
func TestGAR_AC001_NegativeReturnsNotMatched(t *testing.T) {
	requireSG(t)
	rulesDir := stageRulesDir(t)
	root := projectRoot(t)
	fixture := filepath.Join(root, "internal", "astgrep", "testdata", "gar", "negative.go")

	got := findingsFor(t, rulesDir, fixture)
	if len(got) != 0 {
		var locs []string
		for _, f := range got {
			locs = append(locs, fmt.Sprintf("%s:%d", f.File, f.Line))
		}
		t.Errorf("AC-GAR-001: expected 0 matches on negative fixture, got %d: %v", len(got), locs)
	}
}

// TestGAR_AC002_RealErrorReturnStillMatched (AC-GAR-002 / REQ-GAR-002):
// the refined rule MUST still match a genuine unwrapped `return err` inside an
// `if err != nil` block — exactly one match on the positive fixture.
func TestGAR_AC002_RealErrorReturnStillMatched(t *testing.T) {
	requireSG(t)
	rulesDir := stageRulesDir(t)
	root := projectRoot(t)
	fixture := filepath.Join(root, "internal", "astgrep", "testdata", "gar", "positive.go")

	got := findingsFor(t, rulesDir, fixture)
	if len(got) != 1 {
		t.Errorf("AC-GAR-002: expected exactly 1 match on positive fixture, got %d", len(got))
	}
}

// TestGAR_AC003_AutofixDoesNotTouchNegatives (AC-GAR-003 / REQ-GAR-003):
// invoking `sg scan --update-all` with the shipped rule (fix field and all)
// MUST leave the negative fixture byte-identical. The fix fires under the
// SAME guard as the match, so a non-match MUST never be rewritten. The moai
// gate scanner itself never applies autofix (it passes neither --update nor
// --rewrite), so this test guards the rule YAML in isolation as well.
func TestGAR_AC003_AutofixDoesNotTouchNegatives(t *testing.T) {
	requireSG(t)
	rulesDir := stageRulesDir(t)
	root := projectRoot(t)
	src := filepath.Join(root, "internal", "astgrep", "testdata", "gar", "negative.go")

	original, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	work := filepath.Join(t.TempDir(), "negative_copy.go")
	if err := os.WriteFile(work, original, 0o644); err != nil {
		t.Fatalf("write work copy: %v", err)
	}

	// Invoke `sg scan --config <staged>/sgconfig.yml --update-all <work>`.
	// Non-zero exit is tolerated (sg returns non-zero in several report modes).
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sg", "scan",
		"--config", filepath.Join(rulesDir, "sgconfig.yml"),
		"--update-all", work,
	)
	_ = cmd.Run()

	after, err := os.ReadFile(work)
	if err != nil {
		t.Fatalf("read work copy after update: %v", err)
	}
	if string(after) != string(original) {
		t.Errorf("AC-GAR-003: autofix modified the negative fixture (expected byte-identical)\n--- before ---\n%s\n--- after ---\n%s", original, after)
	}
}
