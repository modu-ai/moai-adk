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

// TestGuardBypassMutant_ObserveHomePollution — AC-SA-011 (REQ-SA-011): the
// mutant proof that AC-SA-010's zero is EARNED by the guard, not coincidence.
// The mutant executes newTodoCmd() DIRECTLY, bypassing runTodo's
// liveTodoQueueRootReason gate, in the exact context the historical
// pollution was born in (verdict §A-5): CLAUDE_PROJECT_DIR points at a
// non-git temp directory, so git resolution fails and the queue falls back
// to the home root — one lock + one store under $HOME/.moai/todo. The canary
// HOME is this test's own temp dir through the userHomeDirFn seam; the real
// HOME is never touched.
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

	// THE MUTANT: newTodoCmd().Execute() with no gate — exactly what runTodo
	// refuses to do without todoFixture(t).
	cmd := newTodoCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"add", "t510 mutant probe card"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mutant add failed: %v (stderr: %s)", err, errBuf.String())
	}

	todoRoot := filepath.Join(canaryHome, ".moai", "todo")
	entries, readErr := os.ReadDir(todoRoot)
	if readErr != nil || len(entries) == 0 {
		t.Fatalf("mutant produced NO home pollution under %s (read err=%v) — report the guard boundary this reveals (REQ-SA-011: a pollution-less mutant is a finding, not a pass)",
			todoRoot, readErr)
	}
	t.Logf("mutant pollution observed: %d entr(ies) under canary HOME/.moai/todo — runTodo's liveTodoQueueRootReason gate is what keeps the guarded family at zero",
		len(entries))
}
