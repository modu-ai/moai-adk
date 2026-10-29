package e2e

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/review"
)

// Scenario 5 end to end on lane-19's AuditBudget and lane-18's runner. The limit
// values are test inputs. Persisting the rounds is NOT covered: the claim store
// has no round counter today, so the count lives only in AuditBudget.Rounds and
// is lost with the process (stated in the report as a Gap, not asserted here).

// Rounds are counted when started; the same id again (a lost response) counts
// once; the round after the limit is refused and maps to budget-exhausted.
func TestAuditBudget_CountsRoundsOnce_RefusesPastTheLimit(t *testing.T) {
	b := &review.AuditBudget{Max: 3}
	for _, id := range []string{"r1", "r2", "r2", "r3"} {
		if err := b.Start(id); err != nil {
			t.Fatalf("start %s: %v", id, err)
		}
	}
	if b.Used() != 3 {
		t.Fatalf("used = %d, want 3 (the repeated r2 counts once)", b.Used())
	}
	err := b.Start("r4")
	if !errors.Is(err, review.ErrBudgetExhausted) {
		t.Fatalf("round past the limit: %v, want ErrBudgetExhausted", err)
	}
	if reason, ok := review.BudgetBlock(err); !ok || reason != core.ReasonBudgetExhausted {
		t.Errorf("block reason = %q, %v; want budget-exhausted", reason, ok)
	}
	if err := b.Start("r2"); err != nil {
		t.Errorf("a repeat of an already counted round was refused: %v", err)
	}
	if b.Used() != 3 {
		t.Errorf("a refused round changed the count: %d", b.Used())
	}
}

// A new HEAD never resets the count: the limit is about the Issue's audit
// effort, so a repair that changes the HEAD cannot buy fresh rounds.
func TestAuditBudget_HeadChangeNeverResetsTheCount(t *testing.T) {
	b := &review.AuditBudget{Max: 2}
	_ = b.Start("head-a/round-1")
	_ = b.Start("head-a/round-2")
	if err := b.Start("head-b/round-1"); !errors.Is(err, review.ErrBudgetExhausted) {
		t.Errorf("a round on a new HEAD: %v, want ErrBudgetExhausted", err)
	}
}

// No explicit limit means no rounds, reported as not-configured. No limit is invented.
func TestAuditBudget_Unset_IsNotConfigured_NeverUnlimited(t *testing.T) {
	b := &review.AuditBudget{}
	err := b.Start("r1")
	if !errors.Is(err, review.ErrBudgetUnset) {
		t.Fatalf("unset limit: %v, want ErrBudgetUnset", err)
	}
	if reason, ok := review.BudgetBlock(err); !ok || reason != core.ReasonNotConfigured {
		t.Errorf("block reason = %q, %v; want not-configured", reason, ok)
	}
	if b.Used() != 0 {
		t.Errorf("rounds recorded without a limit: %d", b.Used())
	}
}

// The whole path: rounds run out while the Issue is verifying, the Issue is
// paused with the mapped reason, cannot advance, and resumes only to verifying
// once the operator has acted (the condition is observed by the caller).
func TestAuditBudget_Exhaustion_PausesTheIssue_ThenResumesToVerifying(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateVerifying)
	b := &review.AuditBudget{Max: 1}
	if err := b.Start("round-1"); err != nil {
		t.Fatal(err)
	}
	err := b.Start("round-2")
	reason, ok := review.BudgetBlock(err)
	if !ok {
		t.Fatalf("round 2 did not exhaust the budget: %v", err)
	}
	rec, perr := r.Pause(tk, core.StateBlocked, reason, "operator approves a new audit round or splits the scope", waitOf(tk))
	if perr != nil {
		t.Fatal(perr)
	}
	if rec.State != core.StateBlocked || rec.Reason != core.ReasonBudgetExhausted || rec.Return != core.StateVerifying {
		t.Fatalf("pause record = %+v", rec)
	}
	if _, err := r.Land(tk, headA); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("land while the budget is spent: %v", err)
	}
	if rec, err := r.Resume(tk, false); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("resume without the operator acting: %+v, %v", rec, err)
	}
	if rec, err := r.Resume(tk, true); err != nil || rec.State != core.StateVerifying {
		t.Errorf("resume after the operator acted: %+v, %v", rec, err)
	}
	// Resuming does not refill the budget.
	if err := b.Start("round-3"); !errors.Is(err, review.ErrBudgetExhausted) {
		t.Errorf("round after resume: %v, want ErrBudgetExhausted", err)
	}
}
