package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// The dispatch owner reveals the assigned card and its run — the
// identifiers the operation engine's factory reconciliation keys on (review
// round-21, card t1538). The op's own Target/MissionID stay in the
// supervisor lineage's vocabulary.
func TestGTDOwnerRevealsDispatchReconcileIdentifiers(t *testing.T) {
	o := gtdCLIOwner{reconcileCardID: "t1", reconcileRunID: "run-1"}
	if card, run := o.DispatchReconcileIdentifiers(); card != "t1" || run != "run-1" {
		t.Fatalf("revealed identifiers = (%q, %q), want (t1, run-1)", card, run)
	}
	var caps factory.DispatchIdentifiers = o
	if card, run := caps.DispatchReconcileIdentifiers(); card != "t1" || run != "run-1" {
		t.Fatalf("capability identifiers = (%q, %q), want (t1, run-1)", card, run)
	}
	// An owner that performed no dispatch reveals nothing: the engine
	// falls back to the op's own identifiers.
	empty := gtdCLIOwner{}
	if card, run := empty.DispatchReconcileIdentifiers(); card != "" || run != "" {
		t.Fatalf("empty reveal = (%q, %q), want empty", card, run)
	}
}
