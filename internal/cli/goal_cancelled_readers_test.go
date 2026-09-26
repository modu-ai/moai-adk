package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// writeGoalStatusFixture persists an unmet goal (a failing mechanical
// condition, MaxTurns 0 so the launcher's infinite-goal branch would fire if it
// read the status wrongly) carrying the given status.
func writeGoalStatusFixture(t *testing.T, root, sessionID string, status goal.Status) {
	t.Helper()
	g := goal.NewGoal(sessionID, "cancelled-reader fixture", []goal.Condition{
		{Type: goal.ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 0
	g.Status = status
	if err := goal.SaveGoal(root, g); err != nil {
		t.Fatalf("SaveGoal: %v", err)
	}
}

// TestGoalCancelledStatusReaders covers the internal/cli rows of the
// SPEC-DUAL-HARNESS-HOOK-PARITY-001 design.md §D6 consumer table: each non-test
// reader of goal status handles `cancelled` explicitly, and the stop-goal hook
// surfaces an unrecognised status on stderr instead of blocking on it.
// Non-parallel: t.Setenv mutates process-global state.
func TestGoalCancelledStatusReaders(t *testing.T) {
	t.Run("stop-goal emits nothing for a cancelled goal and leaves it cancelled", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("CLAUDE_PROJECT_DIR", root)
		writeGoalStatusFixture(t, root, "d6-cancelled", goal.StatusCancelled)
		stdout, stderr := driveStopGoalStreams(t, `{"session_id":"d6-cancelled"}`)
		if stdout != "" || stderr != "" {
			t.Fatalf("cancelled goal produced output: stdout=%q stderr=%q", stdout, stderr)
		}
		g, err := goal.LoadGoal(root, "d6-cancelled")
		if err != nil || g == nil {
			t.Fatalf("LoadGoal: g=%v err=%v", g, err)
		}
		if g.Status != goal.StatusCancelled || g.TurnsUsed != 0 {
			t.Fatalf("cancelled goal was evaluated: status=%q turns=%d", g.Status, g.TurnsUsed)
		}
	})

	t.Run("stop-goal surfaces an unrecognised status on stderr without blocking", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("CLAUDE_PROJECT_DIR", root)
		writeGoalStatusFixture(t, root, "d6-unknown", goal.Status("paused"))
		stdout, stderr := driveStopGoalStreams(t, `{"session_id":"d6-unknown"}`)
		if stdout != "" {
			t.Fatalf("unrecognised status produced a stdout decision: %q", stdout)
		}
		if !strings.Contains(stderr, "unrecognised goal status") || !strings.Contains(stderr, `"paused"`) {
			t.Fatalf("stderr does not surface the unrecognised status: %q", stderr)
		}
		g, err := goal.LoadGoal(root, "d6-unknown")
		if err != nil || g == nil {
			t.Fatalf("LoadGoal: g=%v err=%v", g, err)
		}
		if g.Status != goal.Status("paused") {
			t.Fatalf("unrecognised status rewritten to %q", g.Status)
		}
	})

	t.Run("launcher raises the block cap only for an armed goal", func(t *testing.T) {
		clearKanbanLauncherEnv(t)
		root := t.TempDir()
		base := []string{"PATH=/usr/bin"}
		writeGoalStatusFixture(t, root, "d6-armed", goal.StatusArmed)
		writeGoalStatusFixture(t, root, "d6-cancelled", goal.StatusCancelled)
		hasCap := func(env []string) bool {
			for _, e := range env {
				if strings.HasPrefix(e, "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=") {
					return true
				}
			}
			return false
		}
		if !hasCap(injectStopHookBlockCapForGoal(context.Background(), base, root, "d6-armed")) {
			t.Fatal("control: armed infinite goal did not raise the cap")
		}
		if hasCap(injectStopHookBlockCapForGoal(context.Background(), base, root, "d6-cancelled")) {
			t.Fatal("cancelled goal raised the block cap")
		}
	})

	t.Run("handoff save embeds only an armed goal", func(t *testing.T) {
		for _, tc := range []struct {
			status    goal.Status
			wantEmbed bool
		}{{goal.StatusArmed, true}, {goal.StatusCancelled, false}} {
			pd := t.TempDir()
			writeGoalStatusFixture(t, pd, "d6-handoff", tc.status)
			if out, err := runHandoff(t, "", "save", "--project-dir", pd, "--body", "b", "--session", "d6-handoff"); err != nil {
				t.Fatalf("handoff save: %v (%s)", err, out)
			}
			rec, present, err := handoff.ReadPending(pd)
			if err != nil || !present {
				t.Fatalf("read pending: present=%v err=%v", present, err)
			}
			if got := rec.EmbeddedGoal != nil; got != tc.wantEmbed {
				t.Fatalf("status %q: embedded=%v, want %v", tc.status, got, tc.wantEmbed)
			}
		}
	})

	t.Run("goal status renders cancelled verbatim", func(t *testing.T) {
		root := t.TempDir()
		writeGoalStatusFixture(t, root, "d6-cancelled", goal.StatusCancelled)
		var list, human bytes.Buffer
		cmd := &cobra.Command{}
		cmd.SetOut(&list)
		if err := runGoalStatusAll(cmd, root, false); err != nil {
			t.Fatalf("runGoalStatusAll: %v", err)
		}
		if !strings.Contains(list.String(), "[cancelled]") {
			t.Fatalf("status --all does not render cancelled: %q", list.String())
		}
		g, _ := goal.LoadGoal(root, "d6-cancelled")
		cmd.SetOut(&human)
		printGoalHuman(cmd, g)
		if !strings.Contains(human.String(), "status:     cancelled") {
			t.Fatalf("status does not render cancelled: %q", human.String())
		}
	})

	t.Run("goal clear removes a cancelled goal's state and writes no status", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("CLAUDE_PROJECT_DIR", root)
		writeGoalStatusFixture(t, root, "d6-clear", goal.StatusCancelled)
		cmd := &cobra.Command{}
		cmd.SetOut(&bytes.Buffer{})
		if err := runGoalClear(cmd, "d6-clear", false); err != nil {
			t.Fatalf("runGoalClear: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, goalStateDirForTest, "d6-clear.json")); !os.IsNotExist(err) {
			t.Fatalf("goal state survives clear: %v", err)
		}
	})
}
