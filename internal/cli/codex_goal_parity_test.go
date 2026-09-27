package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2f — goal parity (card t1099).
// AC-HPR-012 (golden), AC-HPR-013 (precedence through the real producers), and
// AC-HPR-015 (host override is not success). AC-HPR-014 lives in
// internal/goal (TestGoalBudgetTerminationNotSuccess); its Codex-chain leg is
// TestCodexGoalBudgetTerminationRecorded below.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// goalFixture is the Stop-chain fixture with a HEAD that is not a sync-phase
// commit, so the sync gate is not applicable and the goal member alone
// decides.
func goalFixture(t *testing.T) *stopFixture {
	t.Helper()
	f := newStopFixture(t)
	f.commit(t, "feat: work in progress")
	return f
}

// armParityGoal saves an armed goal with one mechanical condition for session.
func armParityGoal(t *testing.T, root, session, cmd string, maxTurns int) *goal.Goal {
	t.Helper()
	g := &goal.Goal{SessionID: session, Goal: "parity", Status: goal.StatusArmed,
		Conditions: []goal.Condition{{Type: goal.ConditionMechanical, Cmd: cmd}},
		Ceiling:    goal.Ceiling{MaxTurns: maxTurns}, ProgressionMode: goal.ProgressionAutonomous,
		CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := goal.SaveGoal(root, g); err != nil {
		t.Fatal(err)
	}
	return g
}

// recordGoalReceipt records a goal receipt for cmd against the current tree.
func recordGoalReceipt(t *testing.T, root, cmd string, exit int) {
	t.Helper()
	ctx := context.Background()
	key, err := verify.Key(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verify.RecordCheck(root, key, verify.CheckEntry{CheckID: "goal", Command: cmd, ExitCode: exit, RecordedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
}

// codexStop runs one whole Codex Stop chain (member 1 allowing, budgets wide so
// the test decides decisions rather than timing) and returns the rendered
// output and the verdict record.
func codexStop(t *testing.T, root string, in *hook.HookInput) (string, *codexStopChainRecord) {
	t.Helper()
	c := newCodexStopChain(root, in)
	c.member1 = allowMember1
	c.budgetFor = wideStopBudget
	out, err := renderCodexStop(c.run(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	rec, err := readCodexStopChainRecord(root, in.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	return string(out), rec
}

// stopReason decodes the reason of a rendered Codex Stop output.
func stopReason(t *testing.T, out string) string {
	t.Helper()
	var v struct {
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("Codex Stop output is not JSON: %q", out)
	}
	return v.Reason
}

func loadGoalStatus(t *testing.T, root, session string) goal.Status {
	t.Helper()
	g, err := goal.LoadGoal(root, session)
	if err != nil {
		t.Fatal(err)
	}
	if g == nil {
		return ""
	}
	return g.Status
}

// TestCodexGoalContinueUntilMet is the AC-HPR-012 golden leg.
func TestCodexGoalContinueUntilMet(t *testing.T) {
	fakeCodexVersion(t, "codex-cli 0.0.0-golden")
	f := goalFixture(t)
	sentinel := filepath.Join(t.TempDir(), "executed")
	cmd := fmt.Sprintf("touch %s && test -f %s", sentinel, filepath.Join(f.root, "done.flag"))
	armParityGoal(t, f.root, "codex-goal", cmd, 30)

	// 1. unmet, fresh failing receipt → continue with the evaluator's reason.
	recordGoalReceipt(t, f.root, cmd, 1)
	out, rec := codexStop(t, f.root, stopInput("codex-goal", false))
	if !strings.Contains(out, `"decision":"block"`) || !strings.Contains(stopReason(t, out), cmd) {
		t.Fatalf("unmet goal with a failing receipt: Codex output = %s, want a continuation naming the failed condition", out)
	}
	if got := rec.status(3); got != reasonUnmet {
		t.Fatalf("unmet goal recorded %q, want %q", got, reasonUnmet)
	}

	// 2. the tree moves, so no receipt matches → continue naming the command
	// to run (unmeasured), and nothing is executed in the hook.
	f.dirty(t, "package main\n\nfunc main() { _ = 2 }\n")
	out, rec = codexStop(t, f.root, stopInput("codex-goal", false))
	if !strings.Contains(out, `"decision":"block"`) || !strings.Contains(out, codexwiring.GoalReceiptCommand) {
		t.Fatalf("receipt absent: Codex output = %s, want a continuation naming %q", out, codexwiring.GoalReceiptCommand)
	}
	if got := rec.status(3); got != reasonUnmeasured {
		t.Fatalf("receipt-absent goal recorded %q, want %q", got, reasonUnmeasured)
	}
	if _, err := os.Stat(sentinel); err == nil {
		t.Fatal("the Codex Stop chain executed the goal condition; it must only look up a receipt")
	}

	// 3. a fresh passing receipt → the next Stop allows and the goal reads satisfied.
	recordGoalReceipt(t, f.root, cmd, 0)
	out, rec = codexStop(t, f.root, stopInput("codex-goal", false))
	if strings.Contains(out, `"decision":"block"`) {
		t.Fatalf("met goal: Codex output = %s, want allow", out)
	}
	if st := loadGoalStatus(t, f.root, "codex-goal"); st != goal.StatusSatisfied {
		t.Fatalf("met goal status = %q, want %q", st, goal.StatusSatisfied)
	}
	if got := rec.status(3); got != stopStatusPass {
		t.Fatalf("met goal recorded %q, want %q", got, stopStatusPass)
	}
}

// runGoalClearFor runs `moai goal clear --session <session>` against root.
func runGoalClearFor(t *testing.T, root, session string) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := runGoalClear(cmd, session, false); err != nil {
		t.Fatalf("moai goal clear: %v (%s)", err, buf.String())
	}
}

// TestGoalCancellationPrecedence is the AC-HPR-013 precedence leg: both real
// producers — the Codex Interrupt handler and `moai goal clear` — take
// precedence over an unmet goal, and neither reads as satisfied. Neither leg
// injects a pre-written cancellation record.
func TestGoalCancellationPrecedence(t *testing.T) {
	fakeCodexVersion(t, "codex-cli 0.0.0-golden")

	t.Run("interrupt", func(t *testing.T) {
		f := goalFixture(t)
		armParityGoal(t, f.root, "cancel-int", "false", 30)
		recordGoalReceipt(t, f.root, "false", 1)
		if out, _ := codexStop(t, f.root, stopInput("cancel-int", false)); !strings.Contains(out, `"decision":"block"`) {
			t.Fatalf("premise: the unmet goal does not block before the interrupt: %s", out)
		}

		in := &hook.HookInput{HookEventName: "Interrupt", SessionID: "cancel-int"}
		if _, err := runCodexSubcommandAt(t, f.root, "interrupt", "codex", in, &faultRegistry{}); err != nil {
			t.Fatalf("moai hook interrupt --harness codex: %v", err)
		}
		before, err := goal.LoadGoal(f.root, "cancel-int")
		if err != nil || before == nil {
			t.Fatalf("goal after interrupt: %+v (%v)", before, err)
		}

		for i := 1; i <= 2; i++ {
			out, rec := codexStop(t, f.root, stopInput("cancel-int", false))
			if strings.Contains(out, `"decision":"block"`) {
				t.Fatalf("Stop %d after the interrupt blocked: %s", i, out)
			}
			if st := loadGoalStatus(t, f.root, "cancel-int"); st != goal.StatusCancelled {
				t.Fatalf("Stop %d after the interrupt: status = %q, want %q", i, st, goal.StatusCancelled)
			}
			if got := rec.status(3); got == stopStatusPass {
				t.Fatalf("Stop %d: a cancelled goal was recorded %q", i, got)
			}
		}
		after, _ := goal.LoadGoal(f.root, "cancel-int")
		if after.TurnsUsed != before.TurnsUsed {
			t.Fatalf("the loop resumed after the interrupt: turns %d → %d", before.TurnsUsed, after.TurnsUsed)
		}
	})

	t.Run("clear", func(t *testing.T) {
		f := goalFixture(t)
		armParityGoal(t, f.root, "cancel-clr", "false", 30)
		recordGoalReceipt(t, f.root, "false", 1)
		runGoalClearFor(t, f.root, "cancel-clr")

		out, rec := codexStop(t, f.root, stopInput("cancel-clr", false))
		if strings.Contains(out, `"decision":"block"`) {
			t.Fatalf("Stop after moai goal clear blocked: %s", out)
		}
		if st := loadGoalStatus(t, f.root, "cancel-clr"); st != "" {
			t.Fatalf("moai goal clear left goal state with status %q", st)
		}
		if got := rec.status(3); got == stopStatusPass {
			t.Fatalf("a cleared goal was recorded %q", got)
		}
		if v, _ := goal.LoadVerdict(f.root, "cancel-clr"); v != nil {
			t.Fatalf("a cleared goal left a verdict: %+v", v)
		}
	})
}

// TestGoalHostOverrideNotSuccess is AC-HPR-015: when the host stops despite a
// block — a stop_hook_active turn, or the host's consecutive-block cap ending
// the turn — an unmet goal is not satisfied, on either harness path.
func TestGoalHostOverrideNotSuccess(t *testing.T) {
	fakeCodexVersion(t, "codex-cli 0.0.0-golden")

	t.Run("stop_hook_active", func(t *testing.T) {
		f := goalFixture(t)
		armParityGoal(t, f.root, "override-active", "false", 30)
		recordGoalReceipt(t, f.root, "false", 1)
		_, rec := codexStop(t, f.root, stopInput("override-active", true))
		if st := loadGoalStatus(t, f.root, "override-active"); st == goal.StatusSatisfied {
			t.Fatalf("Codex: a stop_hook_active turn marked the unmet goal %q", st)
		}
		if got := rec.status(3); got == stopStatusPass {
			t.Fatalf("Codex: a stop_hook_active turn recorded the unmet goal %q", got)
		}

		armParityGoal(t, f.root, "override-active-claude", "false", 30)
		if _, _, _ = evaluateStopGoal(context.Background(), f.root, "override-active-claude", realCmdRunner{}, nil, os.Stderr); loadGoalStatus(t, f.root, "override-active-claude") == goal.StatusSatisfied {
			t.Fatal("Claude: the unmet goal read satisfied")
		}
	})

	t.Run("consecutive-block cap", func(t *testing.T) {
		// The host ends the turn after its own block cap (Claude Code's
		// CLAUDE_CODE_STOP_HOOK_BLOCK_CAP defaults to 8): MoAI sees 8 blocked
		// Stops and then nothing. The goal must not read satisfied afterwards.
		f := goalFixture(t)
		armParityGoal(t, f.root, "override-cap", "false", 30)
		recordGoalReceipt(t, f.root, "false", 1)
		for i := 0; i < 8; i++ {
			codexStop(t, f.root, stopInput("override-cap", i > 0))
		}
		st := loadGoalStatus(t, f.root, "override-cap")
		g, _ := goal.LoadGoal(f.root, "override-cap")
		if g != nil {
			t.Logf("after 8 host-capped Stops the goal reads %q after %d evaluation(s)", st, g.TurnsUsed)
		}
		if st == goal.StatusSatisfied || st == "" {
			t.Fatalf("after the host's block cap the goal reads %q, want an unsatisfied status", st)
		}
		if v, _ := goal.LoadVerdict(f.root, "override-cap"); v != nil && strings.Contains(strings.ToLower(v.Reason), "satisfied") {
			t.Fatalf("a verdict claims satisfaction: %+v", v)
		}
	})
}

// TestCodexGoalBudgetTerminationRecorded is the Codex-chain leg of AC-HPR-014:
// a goal that reaches its turn ceiling on the Codex path stops the loop, has
// its verdict persisted, and is recorded as terminated — never as a pass.
func TestCodexGoalBudgetTerminationRecorded(t *testing.T) {
	fakeCodexVersion(t, "codex-cli 0.0.0-golden")
	f := goalFixture(t)
	armParityGoal(t, f.root, "budget", "false", 2)
	recordGoalReceipt(t, f.root, "false", 1)

	if out, _ := codexStop(t, f.root, stopInput("budget", false)); !strings.Contains(out, `"decision":"block"`) {
		t.Fatalf("turn 1 below the ceiling did not continue: %s", out)
	}
	out, rec := codexStop(t, f.root, stopInput("budget", false))
	if strings.Contains(out, `"decision":"block"`) {
		t.Fatalf("the turn ceiling did not stop the loop: %s", out)
	}
	if st := loadGoalStatus(t, f.root, "budget"); st != goal.StatusCeilingExit {
		t.Fatalf("status after the ceiling = %q, want %q", st, goal.StatusCeilingExit)
	}
	if v, err := goal.LoadVerdict(f.root, "budget"); err != nil || v == nil {
		t.Fatalf("no verdict persisted at the ceiling: %+v (%v)", v, err)
	}
	if got := rec.status(3); got != stopStatusTerminated {
		t.Fatalf("a goal ended by its budget was recorded %q, want %q", got, stopStatusTerminated)
	}
}
