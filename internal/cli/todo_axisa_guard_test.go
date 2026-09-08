package cli

// todo_axisa_guard_test.go — SPEC-STATE-ANCHOR-001 M4, Axis-A verification
// (AC-SA-010 / AC-SA-011). No guard behavior is changed here (card t422's
// fail-loud repair at e7a078970 stays untouched): this file only PRESERVES
// the evidence — the swept-set liveness guard and the bypass-mutant
// observation — as committed, runnable artifacts.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestTodoSweepSelectorMatchesFamily — AC-SA-010's selector-0-match guard.
// The canary-HOME sweep runs the todo family through `-run TestTodo`; a
// selector matching zero tests reports the same "ok" as a fully passing run,
// so an empty swept set asserts nothing (verification-completeness §1.1).
// The swept set is pinned here: if the todo family is ever renamed away from
// the selector, this guard fires instead of the sweep silently going vacuous.
func TestTodoSweepSelectorMatchesFamily(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	re := regexp.MustCompile(`^func TestTodo\w*\(\s*t\s+\*testing\.T`)
	n := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, readErr := os.ReadFile(e.Name())
		if readErr != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			if re.MatchString(line) {
				n++
			}
		}
	}
	if n == 0 {
		t.Fatalf("selector TestTodo matches 0 test functions in this package — the canary-HOME sweep would be a vacuous green")
	}
	t.Logf("selector TestTodo matches %d test functions (swept-set liveness guard)", n)
}

// TestAxisACanaryHomeSweep_TodoFamily — AC-SA-010 (REQ-SA-010): the
// canary-HOME sweep of the todo family, as a runnable instrument. The whole
// family (`-run TestTodo`) runs as a child `go test` whose HOME is this
// test's own canary temp dir (t.Setenv — the sanctioned isolation path, plan
// D5; the worktree Bash guard rightly refuses a shell-level HOME override).
// The verdict: zero directories under the canary HOME/.moai/todo, with the
// swept count reported — a run whose swept set is empty FAILS here instead
// of reporting a vacuous green.
//
// The name deliberately does NOT match the TestTodo selector (which would
// make the child re-run this very test — infinite regress), and the child's
// env carries the gate flag cleared as a second belt.
//
// Gated behind MOAI_AXIS_A_CANARY_SWEEP=1: the child re-runs the cli package,
// so CI does not pay it by default — it is the on-demand re-measurement of
// this SPEC's M4 evidence. (The always-on liveness pin is
// TestTodoSweepSelectorMatchesFamily above.)
func TestAxisACanaryHomeSweep_TodoFamily(t *testing.T) {
	if os.Getenv("MOAI_AXIS_A_CANARY_SWEEP") != "1" {
		t.Skip("set MOAI_AXIS_A_CANARY_SWEEP=1 to re-run the canary-HOME sweep (runs the todo family as a child go test)")
	}
	canary := t.TempDir()
	// Pin the Go caches to their pre-override locations BEFORE HOME moves:
	// with HOME redirected the go tool would default GOMODCACHE to
	// <canary>/go/pkg/mod and GOCACHE to <canary>/.cache/go-build, and the
	// module cache's read-only files would then break t.TempDir cleanup.
	for _, key := range []string{"GOMODCACHE", "GOCACHE"} {
		out, envErr := exec.Command("go", "env", key).Output()
		if envErr != nil {
			t.Fatalf("go env %s: %v", key, envErr)
		}
		t.Setenv(key, strings.TrimSpace(string(out)))
	}
	t.Setenv("HOME", canary)
	t.Setenv("MOAI_AXIS_A_CANARY_SWEEP", "") // the child must not re-enter this test
	t.Setenv("MOAI_KANBAN", "")
	t.Setenv("MOAI_KANBAN_ID", "")
	t.Setenv("MOAI_KANBAN_LABEL", "")
	t.Setenv("MOAI_KANBAN_LEAD_ADDR", "")
	t.Setenv("MOAI_KANBAN_SETTINGS_INJECTED", "")

	// Locate the repository root (the child needs the package pattern).
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the package dir")
		}
		dir = parent
	}

	cmd := exec.Command("go", "test", "./internal/cli/", "-run", "TestTodo", "-count=1", "-v")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	log := string(out)
	passes := strings.Count(log, "--- PASS")
	if err != nil {
		t.Fatalf("sweep run failed (exit err=%v, %d passing):\n%s", err, passes, tailLines(log, 30))
	}
	if passes == 0 {
		t.Fatalf("sweep matched 0 tests — selector drift; a zero swept set asserts nothing")
	}

	if entries, readErr := os.ReadDir(filepath.Join(canary, ".moai", "todo")); readErr == nil && len(entries) > 0 {
		t.Fatalf("canary HOME polluted: %d entr(ies) under %s — the guarded family regressed",
			len(entries), filepath.Join(canary, ".moai", "todo"))
	}
	t.Logf("sweep verdict: %d todo tests ran under canary HOME %s — 0 directories created under .moai/todo", passes, canary)
}

// tailLines returns at most n trailing lines of s, for bounded failure logs.
func tailLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// TestGuardBypassMutant_ObserveHomePollution — AC-SA-011 (REQ-SA-011),
// transitioned to that requirement's SECOND branch by
// SPEC-TODO-HOME-TEMP-GUARD-001.
//
// History, kept because the transition is the point. The mutant executes
// newTodoCmd() DIRECTLY, bypassing runTodo's liveTodoQueueRootReason gate, in
// the exact context the historical pollution was born in: CLAUDE_PROJECT_DIR
// points at a non-git temp directory, so git resolution fails and the queue
// used to fall back to the home root — one lock + one store under
// $HOME/.moai/todo. Observing that pollution was the evidence that AC-SA-010's
// zero was EARNED by the runTodo gate rather than coincidence.
//
// The mutant no longer produces pollution, and REQ-SA-011 already says what to
// do about that: "a mutant that produces no pollution shall be reported
// alongside the one that does, naming the guard boundary it reveals." This is
// therefore a transition into a branch that requirement already documents, not
// a withdrawal of it — SPEC-STATE-ANCHOR-001 is not modified, and its own
// acceptance record stays a time-fixed observation of the tree that preceded
// this card.
//
// The boundary revealed: t422's runTodo gate lives in the TEST layer, while
// SPEC-TODO-HOME-TEMP-GUARD-001's temporary-origin refusal lives in the
// RESOLVER layer beneath it. Bypassing the upper gate no longer reaches the
// home root, because the lower one never resolves to it. Two different layers;
// the lower one covers the upper one's bypass. The full report — which mutant,
// which boundary, and the control run on the pre-guard tree that separates
// "the guard worked" from "the mutant did not run" — is M3's artifact at
// .moai/reports/t536/guard-boundary.md.
//
// The name deliberately carries GuardBypassMutant and deliberately does NOT
// match the AC-SA-010 sweep selector (TestTodo): the swept family is the
// GUARDED family; this test is the guarded family's bypassed counterfactual.
func TestGuardBypassMutant_ObserveHomePollution(t *testing.T) {
	canaryHome := t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return canaryHome, nil }
	t.Cleanup(func() { userHomeDirFn = orig })

	nonGit := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", nonGit)

	// The precondition that makes the absence attributable: this base IS a
	// temporary origin, so the resolver-layer guard is the thing standing
	// between the mutant and the home root.
	if reason, isTemp := kanban.TempOriginReason(nonGit); !isTemp {
		t.Fatalf("precondition: %q must classify as a temporary origin (reason %q); "+
			"without it, an absence of pollution says nothing about the resolver-layer guard", nonGit, reason)
	}

	// THE MUTANT: newTodoCmd().Execute() with no gate — exactly what runTodo
	// refuses to do without todoFixture(t).
	cmd := newTodoCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"add", "t536 mutant probe card"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mutant add failed: %v (stderr: %s)", err, errBuf.String())
	}

	todoRoot := filepath.Join(canaryHome, ".moai", "todo")
	entries, readErr := os.ReadDir(todoRoot)
	if readErr == nil && len(entries) > 0 {
		t.Fatalf("the mutant produced home pollution under %s (%d entr(ies)) — "+
			"the resolver-layer temporary-origin guard regressed", todoRoot, len(entries))
	}

	// Positive attribution: the run went somewhere, and that somewhere is the
	// substitute root the guard names. Absence of pollution on its own would be
	// satisfied just as well by a command that did nothing at all.
	root := resolveTodoQueueRoot()
	if root != nonGit {
		t.Fatalf("queue root = %q, want the guard's substitute root %q", root, nonGit)
	}
	rec, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root)).Load()
	if err != nil {
		t.Fatalf("load the project-local queue the run continued against: %v", err)
	}
	if len(rec.Items) == 0 {
		t.Fatalf("the mutant add landed nowhere: the project-local queue at %s is empty, so "+
			"the zero-pollution observation cannot be attributed to the guard",
			kanban.BacklogPathForRoot(root))
	}

	t.Logf("REQ-SA-011 second branch: the bypass mutant produced NO pollution under %s. "+
		"Boundary revealed: t422's runTodo gate (test layer) is bypassed, but "+
		"SPEC-TODO-HOME-TEMP-GUARD-001's temporary-origin refusal (resolver layer) never resolves "+
		"to a home root for this base, so the card landed in the project-local queue at %s instead. "+
		"Full report with the pre-guard control run: .moai/reports/t536/guard-boundary.md",
		todoRoot, kanban.BacklogPathForRoot(root))
}
