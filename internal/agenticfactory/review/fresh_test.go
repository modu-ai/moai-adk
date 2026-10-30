package review

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// S46 HANDOFF08: after the HEAD changes, no earlier writer or audit evidence
// is reused; only evidence for the latest HEAD counts.
func TestHandoff08_HeadChangeInvalidatesPriorEvidence(t *testing.T) {
	oldRecs := map[string]Record{
		"code review": reviewRec(t, ProviderClaude, "code-reviewer", "s-r", KindCodeReview, VerdictPass),
		"security":    reviewRec(t, ProviderClaude, "security-reviewer", "s-s", KindSecurityReview, VerdictPass),
		"audit":       completed(t, VerdictPass, headA, "s-audit"),
	}
	for name, rec := range oldRecs {
		if got := rec.Admit(headB, "s-writer"); got != ReasonStale {
			t.Errorf("%s at old HEAD admitted for new HEAD: %q", name, got)
		}
	}
	oldWriter := &core.Evidence{Path: "p", Head: headA}
	if HandoffCurrent(oldWriter, headB) {
		t.Error("writer handoff for the old HEAD reused for the new HEAD")
	}
	if !HandoffCurrent(&core.Evidence{Path: "p", Head: headB}, headB) {
		t.Error("writer handoff for the latest HEAD must count")
	}
	if HandoffCurrent(nil, headB) {
		t.Error("missing handoff must not count")
	}
	// Fresh evidence for the new HEAD is admitted.
	fresh := completed(t, VerdictPass, headB, "s-audit")
	if got := fresh.Admit(headB, "s-writer"); got != ReasonNone {
		t.Errorf("fresh audit at latest HEAD: %q", got)
	}
}

// Audit round budget (S46 Budget): explicit limit, idempotent per round id,
// not reset by a HEAD change.
func TestAuditBudget(t *testing.T) {
	b := &AuditBudget{Max: 2}
	for _, id := range []string{"r1", "r2", "r1"} { // r1 repeated: response lost, same request id
		if err := b.Start(id); err != nil {
			t.Fatalf("Start(%s): %v", id, err)
		}
	}
	if b.Used() != 2 {
		t.Fatalf("Used = %d, want 2", b.Used())
	}
	err := b.Start("r3") // a new HEAD does not reset the budget
	if !errors.Is(err, ErrBudgetExhausted) {
		t.Fatalf("Start(r3) = %v, want ErrBudgetExhausted", err)
	}
	if r, ok := BudgetBlock(err); !ok || r != core.ReasonBudgetExhausted {
		t.Fatalf("BudgetBlock = %q,%v", r, ok)
	}
}

func TestAuditBudget_UnsetAndEmptyID(t *testing.T) {
	for _, max := range []int{0, -1} {
		err := (&AuditBudget{Max: max}).Start("r1")
		if !errors.Is(err, ErrBudgetUnset) {
			t.Fatalf("Max=%d: %v, want ErrBudgetUnset", max, err)
		}
		if r, ok := BudgetBlock(err); !ok || r != core.ReasonNotConfigured {
			t.Fatalf("Max=%d BudgetBlock = %q,%v", max, r, ok)
		}
	}
	if err := (&AuditBudget{Max: 1}).Start(""); err == nil {
		t.Error("empty round id must be refused")
	}
	if _, ok := BudgetBlock(errors.New("other")); ok {
		t.Error("unrelated error mapped to a block")
	}
}
