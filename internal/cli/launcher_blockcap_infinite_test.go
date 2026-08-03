package cli

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
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
