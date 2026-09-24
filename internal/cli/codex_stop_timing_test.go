package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d — AC-HPR-016 timing leg (REQ-HPR-018).
// Each member the Codex Stop handler runs in-hook is timed on the golden
// fixture, doing its real work (receipts present, gates enabled, a dirty
// tree), over repeated runs with a fresh chain each time so every run pays the
// tree-key computation. The observed maximum must not exceed the member's
// declared internal budget (codexwiring.StopChainMembers). The observed
// figures are logged as `stop-timing member=<n> max=<d> budget=<d>` lines for
// the progress record.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// stopTimingRuns is the number of timed runs per member.
const stopTimingRuns = 5

// stopTimingTelemetryRecords sizes the session telemetry member 1 reads, so
// its reflection step runs (it needs at least 3 records) on a realistic file.
const stopTimingTelemetryRecords = 200

// memberBudgetViolation names a member whose observed maximum exceeds its
// declared budget (the uncut step counts toward member 1's allowance).
func memberBudgetViolation(n int, observed time.Duration) string {
	spec := stopMemberSpec(n)
	allowance := spec.Budget + spec.UncutBudget
	if observed > allowance {
		return fmt.Sprintf("member %d (%s) observed max %s exceeds its declared budget %s", n, spec.Name, observed, allowance)
	}
	return ""
}

// newTimingFixture builds a fixture where every member does its real work.
func newTimingFixture(t *testing.T) *stopFixture {
	t.Helper()
	ctx := context.Background()
	fakeCodexVersion(t, "codex-cli 0.0.0-timing")
	withCodexSession(t, codexSessionScript("clean change, approved"))
	t.Setenv(config.EnvSecurityCommitReview, "1")
	f := newStopFixture(t)
	f.enableReviewGates(t, true, true)
	f.write(t, ".moai/config/sections/system.yaml", "hook:\n  opt_in:\n    enabled: true\n")
	f.dirty(t, "reviewable, with a password = \"hunter2\" line for the guardian")
	if _, err := produceSyncGateReceipt(ctx, f.root); err != nil {
		t.Fatal(err)
	}
	if _, err := produceCodexReviewReceipt(ctx, f.root); err != nil {
		t.Fatal(err)
	}
	f.write(t, ".moai/state/audit-multi/timing-s.json", `{"overall_verdict":"pass","residual_risk_note":"timing"}`)
	armGoal(t, f.root, "timing-s", "false", intp(1))

	var tel strings.Builder
	now := time.Now().UTC()
	for i := range stopTimingTelemetryRecords {
		fmt.Fprintf(&tel, `{"ts":%q,"session_id":"timing-s","skill_id":"skill-%d","trigger":"auto","context_hash":"abcd1234","agent_type":"manager-develop","phase":"run","duration_ms":%d,"outcome":"success"}`+"\n",
			now.Add(-time.Duration(i)*time.Second).Format(time.RFC3339), i%7, i)
	}
	f.write(t, ".moai/evolution/telemetry/usage-"+now.Format("2006-01-02")+".jsonl", tel.String())
	return f
}

func TestStopChainMemberCostWithinBudget(t *testing.T) {
	f := newTimingFixture(t)
	input := stopInput("timing-s", false)
	input.CWD = f.root

	stopHandler := hook.NewStopHandler()
	members := map[int]func(ctx context.Context, c *codexStopChain){
		1: func(ctx context.Context, _ *codexStopChain) { _, _ = stopHandler.Handle(ctx, input) },
		2: func(ctx context.Context, c *codexStopChain) { c.syncGateMember(ctx) },
		3: func(ctx context.Context, c *codexStopChain) {
			armGoal(t, f.root, "timing-s", "false", nil) // keep the turn count below the ceiling
			c.goalMember(ctx)
		},
		4: func(ctx context.Context, c *codexStopChain) { _, _ = c.advisory[4](ctx) },
		5: func(ctx context.Context, c *codexStopChain) { _, _ = c.advisory[5](ctx) },
		6: func(ctx context.Context, c *codexStopChain) { c.codexReviewMember(ctx) },
		7: func(ctx context.Context, c *codexStopChain) { c.multiReviewMember(ctx) },
		8: func(ctx context.Context, c *codexStopChain) {
			// Positive control: a skipped observer would time nothing.
			if _, err := c.advisory[8](ctx); errors.Is(err, errAdvisorySkipped) {
				t.Fatal("member 8 skipped its work; the timing probe would measure nothing")
			}
		},
	}

	for _, spec := range codexwiring.StopChainMembers {
		fn := members[spec.Number]
		if fn == nil {
			t.Fatalf("member %d (%s) has no timing probe", spec.Number, spec.Name)
		}
		var observed time.Duration
		for range stopTimingRuns {
			c := newCodexStopChain(f.root, input)
			start := time.Now()
			fn(context.Background(), c)
			if d := time.Since(start); d > observed {
				observed = d
			}
		}
		t.Logf("stop-timing member=%d max=%s budget=%s", spec.Number, observed.Round(time.Millisecond), spec.Budget+spec.UncutBudget)
		if v := memberBudgetViolation(spec.Number, observed); v != "" {
			t.Error(v)
		}
	}

	// The whole chain under the declared budgets, for chain_overhead.
	var chainMax time.Duration
	for range stopTimingRuns {
		armGoal(t, f.root, "timing-s", "false", nil)
		c := newCodexStopChain(f.root, input)
		c.member1 = func(ctx context.Context, in *hook.HookInput) (*hook.HookOutput, error) {
			return stopHandler.Handle(ctx, in)
		}
		start := time.Now()
		c.run(context.Background())
		if d := time.Since(start); d > chainMax {
			chainMax = d
		}
	}
	t.Logf("stop-timing chain max=%s deadline=%s", chainMax.Round(time.Millisecond), newCodexStopChain(f.root, input).stopChainDeadline())

	t.Run("the checker names a member over its budget", func(t *testing.T) {
		over := stopMemberSpec(7).Budget + time.Millisecond
		if v := memberBudgetViolation(7, over); v == "" {
			t.Fatalf("an observed %s against member 7's %s budget was not reported", over, stopMemberSpec(7).Budget)
		}
	})
}
