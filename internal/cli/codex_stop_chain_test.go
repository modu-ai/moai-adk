package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d — the Codex Stop chain beyond the
// per-member parity goldens: AC-HPR-003 (no dependency on .claude/ under the
// gpt profile), AC-HPR-005 (a failed advisory member is recorded and never
// passed, the decision unaffected), and the `moai hook stop --harness codex`
// wiring.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// armGoal arms a one-condition mechanical goal for session and, when exit is
// non-nil, records a fresh receipt of that condition for the current tree.
func armGoal(t *testing.T, root, session, cmd string, exit *int) {
	t.Helper()
	g := &goal.Goal{SessionID: session, Goal: "fixture", Status: goal.StatusArmed,
		Conditions: []goal.Condition{{Type: goal.ConditionMechanical, Cmd: cmd}},
		Ceiling:    goal.Ceiling{MaxTurns: 30}, ProgressionMode: goal.ProgressionAutonomous,
		CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := goal.SaveGoal(root, g); err != nil {
		t.Fatal(err)
	}
	if exit == nil {
		return
	}
	key, err := verify.Key(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verify.RecordCheck(root, key, verify.CheckEntry{CheckID: "goal", Command: cmd, ExitCode: *exit, RecordedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
}

func memberByNumber(t *testing.T, res codexStopChainResult, n int) stopMemberOutcome {
	t.Helper()
	for _, m := range res.Members {
		if m.Number == n {
			return m
		}
	}
	t.Fatalf("member %d missing from the chain result", n)
	return stopMemberOutcome{}
}

// TestStopChainGPTProfileNoClaudeDependency is AC-HPR-003: in a project
// deployed with the gpt profile (no .claude/ tree), the required members and
// the goal member execute on the goal-unmet and gate-failing goldens and each
// returns its own decision.
func TestStopChainGPTProfileNoClaudeDependency(t *testing.T) {
	ctx := context.Background()
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"llm": "gpt"})
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err == nil {
		t.Fatal("premise: the gpt profile deployed a .claude/ tree")
	}
	fakeCodexVersion(t, "codex-cli 0.0.0-gpt")
	f := newStopFixtureAt(t, projectDir)
	f.enableReviewGates(t, true, false)
	f.setFakeGo(t, 1) // the sync gate's vet fails
	f.dirty(t, "reviewable")

	// Receipts for the current tree: a failing sync gate, a failing review.
	if _, err := produceSyncGateReceipt(ctx, f.root); err != nil {
		t.Fatalf("sync gate receipt: %v", err)
	}
	withCodexSession(t, codexSessionScript("- [P1] found issues\n- [P2] more"))
	if _, err := produceCodexReviewReceipt(ctx, f.root); err != nil {
		t.Fatalf("codex review receipt: %v", err)
	}
	armGoal(t, f.root, "gpt-s", "false", intp(1))

	c := newCodexStopChain(f.root, stopInput("gpt-s", false))
	c.member1 = allowMember1
	c.budgetFor = wideStopBudget
	res := c.run(ctx)

	for _, want := range []struct {
		n     int
		class string
	}{
		{2, reasonGateFailed},
		{3, reasonUnmet},
		{6, reasonGateFailed},
	} {
		m := memberByNumber(t, res, want.n)
		if m.Decision != codexadapter.DecisionDeny || m.Class != want.class {
			t.Errorf("member %d (%s): got %s/%q (%s), want deny/%q", want.n, m.Name, m.Decision, m.Class, m.Reason, want.class)
		}
	}
	if res.Output == nil || res.Output.Decision != hook.DecisionBlock {
		t.Fatalf("merged output = %+v, want a Stop block", res.Output)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err == nil {
		t.Fatal("the Codex Stop chain created a .claude/ path in a gpt-profile project")
	}
}

// TestStopChainAdvisoryFailureRecorded is AC-HPR-005: an advisory member that
// fails, or exceeds its internal budget, leaves the decision unchanged, is
// recorded failed, and is never recorded as passed.
func TestStopChainAdvisoryFailureRecorded(t *testing.T) {
	ctx := context.Background()
	f := newStopFixture(t)
	f.write(t, "main.go", "package main\n\nfunc main() { _ = 5 }\n")
	f.commit(t, "feat: not a sync commit")
	armGoal(t, f.root, "adv-s", "false", intp(1))

	chain := func() *codexStopChain {
		c := newCodexStopChain(f.root, stopInput("adv-s", false))
		c.member1 = allowMember1
		return c
	}

	base := chain()
	base.budgetFor = wideStopBudget
	baseRes := base.run(ctx)

	failing := chain()
	// Member 5 below ignores its context and is cut off at 50ms, so its
	// goroutine outlives run(). Join it before this test returns (card t1099).
	defer failing.waitOrphans()
	failing.budgetFor = func(n int) time.Duration {
		if n == 5 {
			return 50 * time.Millisecond
		}
		return wideStopBudget(n)
	}
	failing.advisory[4] = func(context.Context) (string, error) {
		return "", errors.New("injected advisory failure")
	}
	failing.advisory[5] = func(context.Context) (string, error) {
		time.Sleep(time.Second) // ignores its context: a hung member
		return "late", nil
	}
	// A goal counts one turn per evaluation; re-arm (keeping the same
	// receipt) so both runs see turn 1.
	armGoal(t, f.root, "adv-s", "false", nil)
	res := failing.run(ctx)

	if baseRes.Output == nil || res.Output == nil || baseRes.Output.Decision != res.Output.Decision || baseRes.Output.Reason != res.Output.Reason {
		t.Fatalf("an advisory failure changed the decision:\n base   %+v\n failed %+v", baseRes.Output, res.Output)
	}
	rec, err := readCodexStopChainRecord(f.root, "adv-s")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{4, 5} {
		st := rec.status(n)
		if st != stopStatusFailed {
			t.Errorf("member %d recorded %q, want %q", n, st, stopStatusFailed)
		}
		if st == stopStatusPass || st == stopStatusOK {
			t.Errorf("member %d was recorded as passed (%q) although it failed", n, st)
		}
	}
	if m := memberByNumber(t, res, 5); !strings.Contains(m.Err, "budget") {
		t.Errorf("member 5 cut-off error = %q, want it to name the budget", m.Err)
	}
}

// TestStopChainGateCutOffNeverAllows pins design §D3.5: a goal or required
// gate cut off at its internal budget yields `unmeasured` and never allows;
// the fail-open-on-missing member 7 cut off reads as result-missing — it
// allows, with the discard record.
func TestStopChainGateCutOffNeverAllows(t *testing.T) {
	f := newTimingFixture(t)
	c := newCodexStopChain(f.root, stopInput("timing-s", false))
	// Members 2/3/6/7 are cut off below, so their goroutines outlive run()
	// by design. Join them before this test returns: an orphan still reading
	// the package-level seams races with the next test's fixture setup —
	// the intermittent CI Race failure this wait closes (card t1099).
	defer c.waitOrphans()
	c.member1 = allowMember1
	c.budgetFor = func(n int) time.Duration {
		switch n {
		case 2, 3, 6, 7:
			return time.Nanosecond
		}
		return wideStopBudget(n)
	}
	res := c.run(context.Background())
	for _, n := range []int{2, 3, 6} {
		m := memberByNumber(t, res, n)
		if m.Decision != codexadapter.DecisionDeny || m.Class != reasonUnmeasured || !strings.Contains(m.Err, "budget") {
			t.Errorf("member %d cut off: got %s/%q err %q, want a deny/unmeasured naming the budget", n, m.Decision, m.Class, m.Err)
		}
	}
	m7 := memberByNumber(t, res, 7)
	if m7.Decision != codexadapter.DecisionAllow || m7.Status != stopStatusFailOpen || len(m7.Discards) != 1 {
		t.Errorf("member 7 cut off: got %s/%q with %d discard(s), want a fail-open allow with one record", m7.Decision, m7.Status, len(m7.Discards))
	}
	if res.Output == nil || res.Output.Decision != hook.DecisionBlock {
		t.Fatalf("merged output = %+v, want a Stop block", res.Output)
	}
}

// TestStopChainBudgetedCutOffBeatsADeadlineEdgeMember pins the
// budget-enforcement determinism the CI-only flake exposed (card t1293, run
// 36361758033): a member that returns a real verdict AT its deadline — not
// before it — must read as cut off. The runner-observed shape: members 2 and 3
// returned their REAL verdicts (allow, and deny/unmeasured with no budget Err)
// although their 1ns budget had expired, because a select between the result
// channel and mctx.Done() picks randomly when both are ready, and the ctx
// timer's firing can lag a fast member's return on a loaded runner. The
// deadline itself, not channel readiness, must decide.
//
// Reproduction shape: the member sleeps exactly its budget, so its return and
// the timer's firing land in the same instant and both channels are ready at
// the blocked select — against the unfixed select this fails roughly half the
// runs; with the send-time deadline check it is deterministically cut off.
func TestStopChainBudgetedCutOffBeatsADeadlineEdgeMember(t *testing.T) {
	f := newStopFixture(t)
	c := newCodexStopChain(f.root, stopInput("edge-s", false))
	const budget = 100 * time.Millisecond
	c.budgetFor = func(int) time.Duration { return budget }
	res := c.budgeted(context.Background(), 2,
		func(context.Context) stopMemberOutcome {
			time.Sleep(budget) // returns at the deadline, never before it
			return stopMemberOutcome{Decision: codexadapter.DecisionAllow} // a real verdict
		},
		c.cutOffUnmeasured(2, "", ""))
	if res.Decision != codexadapter.DecisionDeny || res.Class != reasonUnmeasured || !strings.Contains(res.Err, "budget") {
		t.Fatalf("a member returning at its deadline took a real verdict: got %s/%q err %q, want the cut-off deny/unmeasured naming the budget", res.Decision, res.Class, res.Err)
	}
}

// runCodexStopSubcommand runs `moai hook stop --harness codex` against root,
// with the registry serving out1 as member 1's output.
func runCodexStopSubcommand(t *testing.T, root string, input *hook.HookInput, out1 *hook.HookOutput) (string, error) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	origDeps := deps
	deps = &Dependencies{
		HookRegistry: &codexHarnessRegistry{output: out1},
		HookProtocol: &codexHarnessProtocol{input: input},
	}
	t.Cleanup(func() { deps = origDeps })
	var sub *cobra.Command
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "stop" {
			sub = cmd
		}
	}
	if sub == nil {
		t.Fatal("hook subcommand stop not found")
	}
	if err := sub.ParseFlags([]string{"--harness", "codex"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sub.Flags().Set("harness", "") })
	sub.SetContext(context.Background())
	var runErr error
	stdout := captureStdoutDuring(t, func() { runErr = sub.RunE(sub, nil) })
	return stdout, runErr
}

// TestCodexStopHandlerRunsTheChain: the single rendered Codex Stop handler
// runs the whole chain, so an unmet goal continues the Codex turn even though
// member 1 (the registry) has no opinion. Before M2d the handler dispatched
// member 1 only and this turn ended.
func TestCodexStopHandlerRunsTheChain(t *testing.T) {
	f := newStopFixture(t)
	f.write(t, "main.go", "package main\n\nfunc main() { _ = 6 }\n")
	f.commit(t, "feat: not a sync commit")
	armGoal(t, f.root, "wire-s", "false", intp(1))

	stdout, err := runCodexStopSubcommand(t, f.root, stopInput("wire-s", false), &hook.HookOutput{})
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if !strings.Contains(stdout, `"decision":"block"`) {
		t.Fatalf("codex Stop stdout = %q, want the goal's block", stdout)
	}
	if _, err := readCodexStopChainRecord(f.root, "wire-s"); err != nil {
		t.Fatalf("the chain left no record: %v", err)
	}
}
