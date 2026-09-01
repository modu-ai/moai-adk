package cli

// gate_summary_cli_test.go — SPEC-GATE-THREE-AXES-001 M1, AC-GTA-005 (card t235).
//
// The gate's execution summary is built in internal/hook/quality and is
// verified there. What is verified HERE is the one thing that package cannot
// see: that `moai gate` actually reaches the user's terminal with it on a
// passing run. runGate prints its output only when the output is non-empty
// (gate.go, the pass branch), so before the summary existed a passing run
// printed nothing at all — the silence this SPEC exists to remove.
//
// The assertion binds to the per-step outcome lines of a named toolchain, NOT
// to output length. "stderr is non-empty" would be satisfied by a banner, a
// version line, or a bare `ok`, and acceptance.md rejects that formulation
// explicitly; the sub-test at the bottom holds the predicate to it.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// goSummaryStepLabels are the steps the Go toolchain configures. Each must
// appear in a passing run's summary carrying exactly one outcome token —
// which of the three it carries is not the subject here (that is AC-GTA-001's,
// verified against the toolchain directly in internal/hook/quality).
//
// golangci-lint is deliberately absent: it is an optional step, so whether it
// is configured at all depends on the machine the suite runs on. Asserting it
// would make this test a probe of the developer's PATH.
var goSummaryStepLabels = []string{"go vet", "go test"}

// summaryOutcomeTokens is the closed outcome set a completed run draws from.
var summaryOutcomeTokens = []string{"executed", "skipped", "disabled"}

func TestGateCmd_PassPathEmitsExecutionSummary(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go is not on PATH, so the Go toolchain's steps cannot run: %v", err)
	}

	dir := t.TempDir()
	// No `go` directive: naming a version above the installed toolchain sends
	// the vet and test steps looking for another toolchain to download.
	writeGateFixtureFile(t, dir, "go.mod", "module t235gatecli\n")
	writeGateFixtureFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	// resolveGateProjectDir reads this; t.Setenv correctly forces the test
	// non-parallel, which also keeps the shared gateCmd writer swap below safe.
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	var stderr bytes.Buffer
	gateCmd.SetErr(&stderr)
	t.Cleanup(func() { gateCmd.SetErr(nil) })

	if err := gateCmd.RunE(gateCmd, nil); err != nil {
		t.Fatalf("moai gate did not exit 0 against a passing project: %v\nstderr:\n%s", err, stderr.String())
	}

	out := stderr.String()
	if strings.TrimSpace(out) == "" {
		t.Fatalf("a passing `moai gate` run emitted nothing; the pass path is still silent")
	}

	rows := parseGateSummaryRows(out)
	if len(rows) == 0 {
		t.Fatalf("stderr carries no per-step summary rows; got:\n%s", out)
	}
	for _, label := range goSummaryStepLabels {
		field, ok := rows[label]
		if !ok {
			t.Errorf("the summary omits the configured step %q; got:\n%s", label, out)
			continue
		}
		if n := countOutcomeTokens(field); n != 1 {
			t.Errorf("step %q outcome field %q carries %d outcome tokens, want exactly 1", label, field, n)
		}
	}

	// The named mutant: any non-empty pass-path string. Held against the same
	// predicate this test asserts with, so the kill is mechanical rather than
	// an argument in a comment.
	t.Run("a bare banner does not satisfy the criterion", func(t *testing.T) {
		for _, mutant := range []string{"ok\n", "moai gate v3.1.3\n", "=== quality gate ===\n"} {
			if rows := parseGateSummaryRows(mutant); len(rows) != 0 {
				t.Errorf("output %q was read as carrying summary rows %v; the criterion would pass a mutant that emits any non-empty string", mutant, rows)
			}
		}
	})
}

// parseGateSummaryRows maps each summary row's step label to its outcome
// field. Rows are the "  - <label>: <outcome>[ — detail]" lines the summary
// renderer writes; anything else on stderr is ignored.
//
// The first ": " is the label separator — a label may itself carry a colon
// ("npm run test:run"), but never a colon followed by a space.
func parseGateSummaryRows(out string) map[string]string {
	const detailSep = " — "

	rows := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		row := strings.TrimPrefix(trimmed, "- ")
		idx := strings.Index(row, ": ")
		if idx < 0 {
			continue
		}
		field := row[idx+len(": "):]
		if sep := strings.Index(field, detailSep); sep >= 0 {
			field = field[:sep]
		}
		rows[row[:idx]] = strings.TrimSpace(field)
	}
	return rows
}

// countOutcomeTokens reports how many of the closed outcome set appear in an
// outcome field. Exactly one is required; zero means the row said nothing,
// and more than one means it said two contradictory things.
func countOutcomeTokens(field string) int {
	n := 0
	for _, tok := range summaryOutcomeTokens {
		if strings.Contains(field, tok) {
			n++
		}
	}
	return n
}

// writeGateFixtureFile lays one file into the fixture project.
func writeGateFixtureFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
