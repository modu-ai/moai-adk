package quality

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAstGrepGateReasons pins the ast-grep gate step's pass-path reason: the
// step still passes when sg is absent, but it now says so instead of returning
// the empty string that a clean scan also returns.
//
// The two halves sit at the frames where each is observable. The gate
// constructs its scanner with a fixed binary name, so only the
// scanner-unavailable branch is reachable end-to-end through the gate; the
// distinctness of the two reason classes is asserted on the constants
// directly. Scanner-level classification of a non-sentinel error is covered by
// TestScannerScan_UnavailableSentinel/non-sentinel.
//
// Non-parallel: the reachable half strips PATH with t.Setenv.
func TestAstGrepGateReasons(t *testing.T) {
	t.Run("unavailable reason reaches the gate step", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		projectDir := t.TempDir()
		// The pure-Go suppression sweep runs before the scan and denies on any
		// unpaired ast-grep-ignore, so the fixture carries none.
		if err := os.WriteFile(filepath.Join(projectDir, "main.go"),
			[]byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
			t.Fatalf("write source: %v", err)
		}

		ok, reason := RunAstGrepGateV2(context.Background(), projectDir, &AstGrepGateConfig{
			Enabled:  true,
			RulesDir: ".moai/config/astgrep-rules",
		})

		if !ok {
			t.Fatalf("RunAstGrepGateV2: want pass when sg is absent, got deny with %q", reason)
		}
		if reason == "" {
			t.Fatal(`RunAstGrepGateV2: want a non-empty pass reason naming the skip, got "" — indistinguishable from a clean scan`)
		}
		if reason != astGrepReasonScannerUnavailable {
			t.Errorf("reason: want the scanner-unavailable constant %q, got %q", astGrepReasonScannerUnavailable, reason)
		}
	})

	t.Run("the two pass reasons are distinct", func(t *testing.T) {
		if astGrepReasonScannerUnavailable == "" {
			t.Error("astGrepReasonScannerUnavailable must be non-empty")
		}
		if astGrepReasonScanDegraded == "" {
			t.Error("astGrepReasonScanDegraded must be non-empty")
		}
		if astGrepReasonScannerUnavailable == astGrepReasonScanDegraded {
			t.Errorf("the scanner-unavailable and degraded-scan reasons collapsed to one string: %q", astGrepReasonScannerUnavailable)
		}
	})
}

// TestAstGrepGateRulesetLoadFailureSkipsDistinctly: a ruleset sg cannot load is
// still fail-open — the gate passes — but the pass says the ruleset failed to
// load. It must not read as a clean scan, which passes with an empty reason.
//
// Non-parallel: it needs the real sg on PATH.
func TestAstGrepGateRulesetLoadFailureSkipsDistinctly(t *testing.T) {
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg CLI is not in PATH — the ruleset load failure needs a real sg to occur")
	}

	projectDir := t.TempDir()
	rulesDir := filepath.Join(projectDir, ".moai", "config", "astgrep-rules")
	goDir := filepath.Join(rulesDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("mkdir rules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - go\n"), 0o644); err != nil {
		t.Fatalf("write sgconfig: %v", err)
	}
	// An unparseable rule document: sg exits non-zero and loads nothing.
	broken := "id: [unclosed\nlanguage: go\nseverity: error\nmessage: broken\nrule:\n  pattern: log.Printf($A, $B)\n"
	if err := os.WriteFile(filepath.Join(goDir, "broken.yml"), []byte(broken), 0o644); err != nil {
		t.Fatalf("write broken rule: %v", err)
	}
	// The suppression sweep runs first and denies on an unpaired ast-grep-ignore,
	// so the fixture carries none.
	if err := os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	ok, reason := RunAstGrepGateV2(context.Background(), projectDir, &AstGrepGateConfig{
		Enabled:  true,
		RulesDir: ".moai/config/astgrep-rules",
	})
	if !ok {
		t.Fatalf("RunAstGrepGateV2: want pass (fail-open) when the ruleset does not load, got deny with %q", reason)
	}
	const wantPrefix = "ast-grep scan skipped: ruleset failed to load ("
	if !strings.HasPrefix(reason, wantPrefix) {
		t.Errorf("reason = %q; want the ruleset-load skip message beginning %q", reason, wantPrefix)
	}
	if !strings.Contains(reason, "Cannot parse rule") {
		t.Errorf("reason = %q; want the first line of the sg error (\"Error: Cannot parse rule ...\") inside the skip message", reason)
	}
}
