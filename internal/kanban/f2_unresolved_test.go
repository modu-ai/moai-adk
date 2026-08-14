// f2_unresolved_test.go — RED probe for review finding F2: the unresolved
// verdict must key on the SOURCE (a resolution that found no single tree),
// not on the status STRING. A frontmatter literally carrying
// `status: unresolved` resolves cleanly to a source; it must reach the
// compatibility table and be reported as the pairing violation it is, never
// misfiled as a resolution failure with a false "worktree reports no branch"
// detail.
package kanban

import (
	"strings"
	"testing"
)

// TestReconcileCard_UnresolvedLiteralIsInconsistent — F2: a status literally
// equal to "unresolved" whose SOURCE resolved cleanly (here the primary
// checkout) reaches the table and is reported INCONSISTENT (the literal is
// outside every row), NOT unresolved. The verdict keys on Source.
func TestReconcileCard_UnresolvedLiteralIsInconsistent(t *testing.T) {
	t.Parallel()
	card := Card{SpecID: "SPEC-F2-001", Column: ColumnPlan, LastMovedAt: "t0"}
	cs := &CardStatus{Status: StatusUnresolved, Source: StatusSourcePrimary, SpecFilePresent: true}

	view := ReconcileCard(card, cs)
	if view.Unresolved {
		t.Fatalf("Unresolved = true for a cleanly-resolved primary source — the verdict must key on Source, not on the status string")
	}
	if !view.Inconsistent {
		t.Fatalf("Inconsistent = false; the literal %q is outside every table row and must be the pairing violation, not a resolution failure", StatusUnresolved)
	}
	if view.Dispatchable {
		t.Fatal("an inconsistent card is dispatchable")
	}
	// The detail must not claim a worktree observation that never failed.
	if strings.Contains(view.Details, "worktree") && strings.Contains(view.Details, "no branch") {
		t.Fatalf("detail claims a worktree failure that did not occur: %q", view.Details)
	}
}

// TestReconcileCard_GenuineUnresolvedSourceStaysUnresolved — F2 positive
// control: when the SOURCE is genuinely unresolved (detached worktree), the
// verdict stays Unresolved and does NOT reach the table.
func TestReconcileCard_GenuineUnresolvedSourceStaysUnresolved(t *testing.T) {
	t.Parallel()
	card := Card{SpecID: "SPEC-F2-002", Column: ColumnRun, LastMovedAt: "t0"}
	cs := &CardStatus{Status: StatusUnresolved, Source: StatusSourceUnresolved, SpecFilePresent: true, WorktreePath: "/wt/SPEC-F2-002"}

	view := ReconcileCard(card, cs)
	if !view.Unresolved {
		t.Fatal("genuine unresolved source must stay Unresolved")
	}
	if view.Inconsistent {
		t.Fatal("unresolved source must not be reported inconsistent (distinct outcome, REQ-KB-024 ordering)")
	}
}
