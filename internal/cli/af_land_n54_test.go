package cli

// N54 (audit of 46dbcce0d4 N40, repeated by the audit of 1c05083f71): Enqueue refuses a commit whose newest green, gate
// or review run failed, and finishing a card refuses the same, but `af land` looked only at the card's position, so a
// run that failed while the row waited was landed all the same: the commit went into the trunk, where nothing takes it
// out again. `af land` now asks the ledger's own rule before it lands, with the way out in the message: measure the
// commit again, with a pass.

import (
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afN54Waiting is a card whose c1 is measured (green and gate pass) and enqueued: the row is waiting, the card in MERGING.
func afN54Waiting(t *testing.T) afN3Card {
	t.Helper()
	a := afN38Card(t) // measured through a verb, the ledger bound
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", a.repo)
	if got := a.row().State; got != ledger.QueueWaiting {
		t.Fatalf("setup: the row is %s, want waiting", got)
	}
	return a
}

func TestAFLand_N54AFailingNewestRunWhileTheRowWaitsRefusesTheLanding(t *testing.T) {
	afSkipWithoutSh(t)
	for _, tc := range []struct {
		name     string
		failing  []string
		wantKind string // the first failing kind, in the order the ledger asks: green, gate, review
		repair   func(a afN3Card)
	}{
		{"the green fails", []string{"green"}, "green", func(a afN3Card) { a.measure(a.c1) }},
		{"the gate fails", []string{"gate"}, "gate", func(a afN3Card) { a.measure(a.c1) }},
		{"the review fails", []string{"review"}, "review", func(a afN3Card) { a.record("review", a.c1, "pass") }},
		// the audit's reproduction: the review and then the gate fail while the row waits
		{"the review and the gate fail", []string{"review", "gate"}, "gate", func(a afN3Card) { a.measure(a.c1); a.record("review", a.c1, "pass") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := afN54Waiting(t)
			for _, kind := range tc.failing {
				a.record(kind, a.c1, "fail")
			}
			before, rowBefore := afSnapshot(t, a.st), a.row()
			out, err := a.try("land", "t1", "--gate", "true")
			if err == nil {
				t.Fatalf("landed with a failing newest %s run: %.300s", tc.wantKind, out)
			}
			for _, want := range []string{"newest " + tc.wantKind + " run", "measure", "again"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal does not say %q: %v", want, err)
				}
			}
			if a.inTrunk(a.c1) {
				t.Error("the refused land put c1 in the trunk")
			}
			if got := a.row(); !reflect.DeepEqual(got, rowBefore) {
				t.Errorf("the refused land changed the row: %+v -> %+v", rowBefore, got)
			}
			if after := afSnapshot(t, a.st); !reflect.DeepEqual(before, after) {
				t.Errorf("the refusal changed the ledger:\nbefore %+v\nafter  %+v", before, after)
			}
			// the way out the message names: measure again, with a pass, and the same card lands and finishes
			tc.repair(a)
			if out := a.verb("land", "t1", "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
				t.Fatalf("the land after measuring again: %s", out)
			}
			if !a.inTrunk(a.c1) {
				t.Error("c1 landed and is not in the trunk")
			}
			if err := a.tryFinish(); err != nil {
				t.Fatalf("the card is refused after the landing: %v", err)
			}
			if got := a.stage(); got != core.StageDone {
				t.Errorf("stage %s, want DONE", got)
			}
		})
	}
}

// The control (D5): a passing gate about another commit while the card stays in MERGING obliges nothing, so the landing
// goes through.
func TestAFLand_N54APassingGateOfAnotherCommitDoesNotBlockTheLanding(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN54Waiting(t)
	c2 := a.newCommit()
	a.record("gate", c2, "pass")
	if out := a.verb("land", "t1", "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("the landing was refused after a passing gate of another commit: %s", out)
	}
	if !a.inTrunk(a.c1) {
		t.Error("c1 landed and is not in the trunk")
	}
}
