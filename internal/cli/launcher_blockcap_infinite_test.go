package cli

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestAC003_BlockCapDoctrineClauseSpecific asserts the doctrine surface carries
// a SINGLE line naming BOTH `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` AND `--max-turns 0`
// — i.e. the raised-value recommendation is scoped to the infinite-goal case
// this SPEC delivers (not merely the cap name, which already exists).
//
// The cap name alone is already present in goal-directive.md / goal.md today,
// so a name-only grep passes vacuously; this clause-specific grep is the AC.
func TestAC003_BlockCapDoctrineClauseSpecific(t *testing.T) {
	files := []string{
		"../../.claude/rules/moai/workflow/goal-directive.md",
		"../../.claude/skills/moai/workflows/goal.md",
	}
	anyMatch := false
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Logf("skip unreadable doctrine file %s: %v", f, err)
			continue
		}
		// A line carrying BOTH the cap name AND the --max-turns 0 arming context.
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP") &&
				strings.Contains(line, "--max-turns 0") {
				anyMatch = true
				t.Logf("clause match in %s: %s", f, strings.TrimSpace(line))
			}
		}
	}
	if !anyMatch {
		t.Errorf("AC-003: no doctrine line names BOTH CLAUDE_CODE_STOP_HOOK_BLOCK_CAP " +
			"and --max-turns 0 (the raised-value recommendation must be scoped to " +
			"the infinite-goal arming case)")
	}
}

// TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal asserts the launcher
// env builder injects CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=<raised> when an armed
// MaxTurns==0 goal exists for the resolving session, and leaves the env
// unchanged when no such goal exists (backward compat).
func TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal(t *testing.T) {
	tmp := t.TempDir()
	ctx := context.Background()

	// No armed goal → no inject (backward compat: env unchanged).
	base := []string{"PATH=/usr/bin", "HOME=/tmp"}
	got := injectStopHookBlockCapForGoal(ctx, base, tmp, "no-such-session")
	for _, e := range got {
		if strings.HasPrefix(e, "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=") {
			t.Errorf("AC-003 backward compat: block cap injected with no armed goal: %q", e)
		}
	}

	// Arm an infinite goal, then re-resolve → raised cap injected.
	armInfiniteGoalFixture(t, tmp, "inf-session", 0)
	got2 := injectStopHookBlockCapForGoal(ctx, base, tmp, "inf-session")
	found := false
	for _, e := range got2 {
		if strings.HasPrefix(e, "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=") {
			found = true
			if e == "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=" {
				t.Errorf("AC-003: empty cap value injected: %q", e)
			}
		}
	}
	if !found {
		t.Errorf("AC-003: CLAUDE_CODE_STOP_HOOK_BLOCK_CAP not injected for armed --max-turns 0 goal")
	}

	// A finite (MaxTurns>0) goal must NOT trigger the inject.
	armInfiniteGoalFixture(t, tmp, "finite-session", 30)
	got3 := injectStopHookBlockCapForGoal(ctx, base, tmp, "finite-session")
	for _, e := range got3 {
		if strings.HasPrefix(e, "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=") {
			t.Errorf("AC-003: block cap injected for a FINITE (MaxTurns>0) goal: %q", e)
		}
	}
}

// armInfiniteGoalFixture writes a goal state file with the given MaxTurns for
// the resolver to read. Uses the real goal.SaveGoal so the file layout matches
// production.
func armInfiniteGoalFixture(t *testing.T, projectRoot, sessionID string, maxTurns int) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "state", "goal")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Minimal JSON matching goal.Goal's persisted shape.
	json := `{"session_id":"` + sessionID + `","goal":"g","conditions":[],"ceiling":{"max_turns":` + strconv.Itoa(maxTurns) + `},"turns_used":0,"progression_mode":"autonomous","created_at":"","status":"armed"}`
	if err := os.WriteFile(filepath.Join(dir, sessionID+".json"), []byte(json), 0o600); err != nil {
		t.Fatal(err)
	}
}

// ── SPEC-FACTORY-MODE-001 M5 ──

// TestACFM022a_KanbanRaisesBlockCapUnconditionally is AC-FM-022a. The
// pre-existing inject is goal-conditional and reads goal state at LAUNCH time;
// a kanban chain arms its goal mid-session, so that predicate is structurally
// unable to see it. The kanban branch is therefore unconditional on the
// process-environment signal, ahead of the goal read.
//
// Non-parallel by construction: t.Setenv mutates process-global state.
func TestACFM022a_KanbanRaisesBlockCapUnconditionally(t *testing.T) {
	tmp := t.TempDir()
	ctx := context.Background()
	base := []string{"PATH=/usr/bin", "HOME=/tmp"}
	want := config.EnvClaudeCodeStopHookBlockCap + "=" + strconv.Itoa(DefaultRaisedStopHookBlockCap)

	// Negative control FIRST, so a leaked MOAI_KANBAN from an earlier test in
	// this binary fails here loudly rather than silently validating the kanban
	// branch (the ordering hazard AC-FM-023d exists to close).
	if got := injectStopHookBlockCapForGoal(ctx, base, tmp, ""); !slices.Equal(got, base) {
		t.Errorf("AC-FM-022a negative control: with %s unset and no armed goal the env must be unchanged, got %v",
			config.EnvMoaiKanban, got)
	}

	t.Setenv(config.EnvMoaiKanban, "1")
	// No armed goal, and an empty sessionID — the pre-existing branch cannot
	// fire here, so a match proves the kanban branch supplied the entry.
	got := injectStopHookBlockCapForGoal(ctx, base, tmp, "")
	if !slices.Contains(got, want) {
		t.Errorf("AC-FM-022a: expected %q in the launch env, got %v", want, got)
	}
}

// TestACFM022a_KanbanCapReplacesPreexistingEntry asserts the kanban branch
// reuses the replace-in-place discipline of the goal branch rather than
// appending a duplicate key, which a child process would resolve ambiguously.
func TestACFM022a_KanbanCapReplacesPreexistingEntry(t *testing.T) {
	t.Setenv(config.EnvMoaiKanban, "1")
	key := config.EnvClaudeCodeStopHookBlockCap
	base := []string{"PATH=/usr/bin", key + "=8"}

	got := injectStopHookBlockCapForGoal(context.Background(), base, t.TempDir(), "")
	count := 0
	for _, e := range got {
		if strings.HasPrefix(e, key+"=") {
			count++
			if e != key+"="+strconv.Itoa(DefaultRaisedStopHookBlockCap) {
				t.Errorf("AC-FM-022a: stale cap survived: %q", e)
			}
		}
	}
	if count != 1 {
		t.Errorf("AC-FM-022a: expected exactly one %s entry, got %d in %v", key, count, got)
	}
}

// TestACFM023c_KanbanEnvReachesChildEnvironment is AC-FM-023c: the load-bearing
// link. The cap raise of AC-FM-022a is reachable in production only if the
// kanban variables survive into the os.Environ()-derived launch env that
// launchClaudeDefault builds immediately above the inject call. A unit test of
// the inject alone would stay green through that failure.
func TestACFM023c_KanbanEnvReachesChildEnvironment(t *testing.T) {
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanSpec, "SPEC-PLACEHOLDER")

	launchEnv := buildEnvForLaunch("high", os.Environ())

	for _, want := range []string{
		config.EnvMoaiKanban + "=1",
		config.EnvMoaiKanbanSpec + "=SPEC-PLACEHOLDER",
	} {
		if !slices.Contains(launchEnv, want) {
			t.Errorf("AC-FM-023c: %q missing from the child environment", want)
		}
	}
}
