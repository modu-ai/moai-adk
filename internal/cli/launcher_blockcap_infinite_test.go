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

// clearFactoryLauncherEnv unsets every kanban signal variable plus the runtime
// block-cap key so the inject's negative controls below start from a
// known-absent state. A session running these tests inside Kanban Mode carries
// the launcher-injected MOAI_KANBAN* variables in its ambient env, and the
// inject's kanban branch is unconditional on them — without this isolation the
// "no signal → env unchanged" controls fail on the developer's own machine.
// t.Setenv registers the restore, so the process env is returned to its prior
// value when the test ends. Same pattern as clearFactoryEnv in
// internal/hook/session_start_env_helper_test.go. The three retired markers
// stay in the list by their written-out names, so a surviving session's
// ambient value cannot reach TestRetiredChainSignalsDoNotRaiseBlockCap.
func clearFactoryLauncherEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		retiredLeaderMarker,
		config.EnvFactoryRunID,
		retiredSpecMarker,
		retiredLaneLabelMarker,
		config.EnvFactorySettingsInjected,
		config.EnvFactoryLeadAddr,
		config.EnvMoaiFactoryWorkers,
		config.EnvMoaiFactoryWorker,
		config.EnvClaudeCodeStopHookBlockCap,
	} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

// TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal asserts the launcher
// env builder injects CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=<raised> when an armed
// MaxTurns==0 goal exists for the resolving session, and leaves the env
// unchanged when no such goal exists (backward compat).
func TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal(t *testing.T) {
	clearFactoryLauncherEnv(t)
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

// ── Factory block-cap clause (SPEC-FACTORY-MODE-001 M5; the kanban clause was
// removed by SPEC-LAUNCHER-ENTRY-FLAGS-001 M5a) ──

// TestRetiredChainSignalsDoNotRaiseBlockCap pins the removal of the kanban
// clause: the retired chain and companion signals, set in the environment with
// no armed goal, no longer raise the Stop-hook block cap. The names are written
// out because no launcher publishes them any more.
//
// Non-parallel by construction: t.Setenv mutates process-global state.
func TestRetiredChainSignalsDoNotRaiseBlockCap(t *testing.T) {
	clearFactoryLauncherEnv(t)
	tmp := t.TempDir()
	ctx := context.Background()
	base := []string{"PATH=/usr/bin", "HOME=/tmp"}

	t.Setenv("MOAI_KANBAN", "1")
	t.Setenv(retiredLaneLabelMarker, "run-tjlgt1")
	if got := injectStopHookBlockCapForGoal(ctx, base, tmp, ""); !slices.Equal(got, base) {
		t.Errorf("the retired chain signals raised the block cap: %v, want the base environment unchanged", got)
	}
}

// TestFactoryCapReplacesPreexistingEntry asserts the factory branch reuses the
// replace-in-place discipline of the goal branch rather than appending a
// duplicate key, which a child process would resolve ambiguously.
func TestFactoryCapReplacesPreexistingEntry(t *testing.T) {
	clearFactoryLauncherEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	key := config.EnvClaudeCodeStopHookBlockCap
	base := []string{"PATH=/usr/bin", key + "=8"}

	got := injectStopHookBlockCapForGoal(context.Background(), base, t.TempDir(), "")
	count := 0
	for _, e := range got {
		if strings.HasPrefix(e, key+"=") {
			count++
			if e != key+"="+strconv.Itoa(DefaultRaisedStopHookBlockCap) {
				t.Errorf("stale cap survived: %q", e)
			}
		}
	}
	if count != 1 {
		t.Errorf("expected exactly one %s entry, got %d in %v", key, count, got)
	}
}

// TestFactoryRaisesBlockCap asserts the factory clause of the unconditional
// raise (SPEC-FACTORY-WORKER-FANOUT-001): a session signalled by
// MOAI_FACTORY_WORKERS — leader or lane, both branches set it — takes the
// raised cap, because a factory run's dispatch-driven turn chains are long and
// meant to survive unattended. The goal read cannot see them, since the session
// arms its goal mid-session.
func TestFactoryRaisesBlockCap(t *testing.T) {
	clearFactoryLauncherEnv(t)
	tmp := t.TempDir()
	ctx := context.Background()
	base := []string{"PATH=/usr/bin", "HOME=/tmp"}
	want := config.EnvClaudeCodeStopHookBlockCap + "=" + strconv.Itoa(DefaultRaisedStopHookBlockCap)

	// Negative control: the factory variable alone must gate the branch (the
	// worker-label variable deliberately does not).
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
	if got := injectStopHookBlockCapForGoal(ctx, base, tmp, ""); !slices.Equal(got, base) {
		t.Errorf("negative control: MOAI_FACTORY_WORKER alone must not raise the cap, got %v", got)
	}

	t.Setenv(config.EnvMoaiFactoryWorkers, "4")
	got := injectStopHookBlockCapForGoal(ctx, base, tmp, "")
	if !slices.Contains(got, want) {
		t.Errorf("expected %q in the launch env for a factory session, got %v", want, got)
	}
}

// TestFactoryEnvReachesChildEnvironment: the load-bearing link of the factory
// cap raise. It is reachable in production only if the factory variables
// survive into the os.Environ()-derived launch env that launchClaudeDefault
// builds immediately above the inject call. A unit test of the inject alone
// would stay green through that failure.
func TestFactoryEnvReachesChildEnvironment(t *testing.T) {
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	t.Setenv(config.EnvFactoryRunID, "run-placeholder")

	// Neither the plain Claude path (card t595) nor the gateway path (card t668)
	// wraps os.Environ() any more — both moved their CLAUDE_CODE_EFFORT_LEVEL
	// injection into the --settings payload, because that variable is an
	// override Claude Code refuses to let /effort or /model change mid-session.
	// The one wrapper that still filters the inherited environment, and so still
	// carries the drop hazard this test guards, is buildEnvForGLMLaunch. The
	// assertion follows it rather than becoming a tautology over os.Environ().
	launchEnv := buildEnvForGLMLaunch(config.GLMModels{}, "", "high", os.Environ())

	for _, want := range []string{
		config.EnvMoaiFactoryWorkers + "=1",
		config.EnvFactoryRunID + "=run-placeholder",
	} {
		if !slices.Contains(launchEnv, want) {
			t.Errorf("%q missing from the child environment", want)
		}
	}
}
