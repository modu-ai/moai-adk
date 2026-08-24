package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// t220: handle-stop-goal.sh resolved the moai binary across three tiers with
// three separate `if` blocks, each ending in
//
//	printf '%s' "$INPUT" | exec <bin> hook stop-goal
//
// `exec` on the right of a pipeline replaces the pipeline's SUBSHELL, not the
// wrapper. The wrapper therefore survived the first tier and fell through to
// the next `if`, firing the evaluator again for every tier whose binary
// existed — and `$HOME/go/bin/moai` is the standard Go install location, so a
// second firing was the norm, not an edge case. Consequences, worst first: the
// Stop JSON is emitted twice from one hook entry (double block decision);
// turns_used double-increments, exhausting the ceiling in half the intended
// turns; every mechanical condition runs twice.
//
// TestAC001_GoalAbsentSkipsMoaiBinary already asserts "exactly 1" — it never
// caught this because it stubs only the PATH tier, leaving $HOME/go/bin empty
// so the fall-through had nothing to find. These tests populate EVERY tier,
// which is what makes the count observable.

// stopGoalWrapperCopies returns the repo-relative paths of the three shipped
// copies of the wrapper. They are byte-identical by contract (CLAUDE.local.md
// §2.3 — the .tmpl is the one that actually deploys, so a fix applied to only
// one copy is reverted by the next `moai update`). Running all three is also
// the parity lock.
func stopGoalWrapperCopies() []string {
	return []string{
		".claude/hooks/moai/handle-stop-goal.sh",
		"internal/template/templates/.claude/hooks/moai/handle-stop-goal.sh",
		"internal/template/templates/.claude/hooks/moai/handle-stop-goal.sh.tmpl",
	}
}

// writeLabeledStub writes a stub binary that appends its label to counterPath.
// It drains stdin first so the wrapper's `printf | stub` pipeline never sees
// EPIPE, which would otherwise perturb what is being measured.
func writeLabeledStub(t *testing.T, path, label, counterPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "#!/bin/bash\ncat >/dev/null\necho \"" + label + "\" >> \"" + counterPath + "\"\nexit 0\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write stub %s: %v", path, err)
	}
}

// readCounterLabels returns the labels recorded by the stubs, in order.
func readCounterLabels(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read counter %s: %v", path, err)
	}
	var labels []string
	for _, ln := range strings.Split(string(b), "\n") {
		if s := strings.TrimSpace(ln); s != "" {
			labels = append(labels, s)
		}
	}
	return labels
}

// TestStopGoalWrapperFiresEvaluatorExactlyOnce is execution-based, not a source
// grep: every resolution tier gets a counting stub, an armed goal makes the
// wrapper proceed past its precondition, and the counter must record exactly
// one invocation — the first tier.
func TestStopGoalWrapperFiresEvaluatorExactlyOnce(t *testing.T) {
	requireBash(t)
	requireGit(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Join(cwd, "..", "..")

	for _, rel := range stopGoalWrapperCopies() {
		t.Run(rel, func(t *testing.T) {
			tmp := t.TempDir()
			stubBin := filepath.Join(tmp, "bin")
			homeDir := filepath.Join(tmp, "home")
			counter := filepath.Join(tmp, "invocations.log")

			// Every tier the wrapper knows about is populated, each with a
			// label naming the tier that fired.
			writeLabeledStub(t, filepath.Join(stubBin, "moai"), "tier1-PATH", counter)
			writeLabeledStub(t, filepath.Join(homeDir, "go", "bin", "moai"), "tier2-HOME-GO-BIN", counter)
			writeLabeledStub(t, filepath.Join(homeDir, ".local", "bin", "moai"), "tier3-HOME-LOCAL-BIN", counter)

			// Arm a goal so the wrapper's shell precondition lets it through.
			sessionID := "sess-t220-single-exec"
			stateDir := filepath.Join(repoRoot, ".moai", "state", "goal")
			if err := os.MkdirAll(stateDir, 0o755); err != nil {
				t.Fatal(err)
			}
			stateFile := filepath.Join(stateDir, sessionID+".json")
			if err := os.WriteFile(stateFile, []byte(`{"armed":true}`), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Remove(stateFile) })

			stdin := `{"session_id":"` + sessionID + `","last_assistant_message":"done"}`
			_ = runHookScript(t, rel, stdin, stubBin, homeDir, repoRoot)

			labels := readCounterLabels(t, counter)
			if len(labels) != 1 {
				t.Fatalf("%s invoked the goal evaluator %d time(s) %v; want exactly 1 — a pipeline's `exec` replaces the subshell, not the wrapper, so the wrapper falls through to the next resolution tier",
					rel, len(labels), labels)
			}
			if labels[0] != "tier1-PATH" {
				t.Errorf("%s resolved to %q; want the first tier (PATH)", rel, labels[0])
			}
		})
	}
}

// TestStopGoalWrapperCopiesStayIdentical locks the Template-First contract: the
// deployed copy is the .tmpl, so a fix landing on only one copy is silently
// reverted by the next `moai update`.
func TestStopGoalWrapperCopiesStayIdentical(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Join(cwd, "..", "..")

	copies := stopGoalWrapperCopies()
	first, err := os.ReadFile(filepath.Join(repoRoot, copies[0]))
	if err != nil {
		t.Fatalf("read %s: %v", copies[0], err)
	}
	for _, rel := range copies[1:] {
		b, err := os.ReadFile(filepath.Join(repoRoot, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if string(b) != string(first) {
			t.Errorf("%s differs from %s — the three copies must stay byte-identical, or `moai update` reverts the deployed one", rel, copies[0])
		}
	}
}
