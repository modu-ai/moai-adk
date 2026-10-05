// todo_classify_regression_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M5
// interaction regressions: the merged lead-side serial cycle (SPEC-MANAGER-
// TODO-001) inherits the sorted queue WITHOUT change (constraint C2), and
// the hold state (SPEC-TODO-HOLD-STATE-001) stays invisible to every machine
// selector while sorting merely positions it (constraint C3).
package cli

import (
	"fmt"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func clsFor(prio string) *factory.CardClassification {
	return &factory.CardClassification{Priority: prio, Blocked: false, Mode: factory.ClassModeSerial, Decider: factory.DeciderIdentityLLM}
}

// TestAutoCycleInheritsSortedQueueOrder — the lead-side cycle picks targets
// in queue order; once the queue is kept sorted by classification (the add
// path's locked write), the cycle's FIRST queued target is the
// highest-ranked card with zero changes to the cycle itself.
func TestAutoCycleInheritsSortedQueueOrder(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := store.Add("low first"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Add("high second"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].Classification = clsFor(factory.ClassPriorityLow)
		rec.Items[1].Classification = clsFor(factory.ClassPriorityHigh)
		rec.SortByClassification() // the add path's write shape
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	targets, _, err := autoPickTargets(rec, newAutoLiveness(), root)
	if err != nil {
		t.Fatalf("autoPickTargets: %v", err)
	}
	if len(targets) < 2 {
		t.Fatalf("targets = %d, want at least the two queued cards", len(targets))
	}
	if targets[0].ID != "t2" {
		t.Errorf("first auto target = %s, want t2 (the high card the sort put first — the cycle consumes queue order, which is now priority order)", targets[0].ID)
	}
}

// TestHoldCardInvisibleToSelectors — constraint C3: a held card is invisible
// to the auto scan by its state (positive enumeration), while the sort only
// positions it; the factory auto-promotion arm skips it the same way.
func TestHoldCardInvisibleToSelectors(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := store.Add("normal card"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Add("held card"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].Classification = clsFor(factory.ClassPriorityLow)
		rec.Items[1].Classification = clsFor(factory.ClassPriorityHigh)
		rec.Items[1].State = factory.BacklogStateHold
		rec.SortByClassification()
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	// The sort positions the held high card first...
	if rec.Items[0].ID != "t2" {
		t.Fatalf("held high card did not take the front: %v", rec.Items)
	}
	// ...and the auto scan still sees only the queued card.
	targets, _, err := autoPickTargets(rec, newAutoLiveness(), root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tgt := range targets {
		if tgt.ID == "t2" {
			t.Errorf("held card t2 appeared among auto targets: %v", targets)
		}
	}
	if got := fmt.Sprint(func() []string {
		var ids []string
		for _, tgt := range targets {
			ids = append(ids, tgt.ID)
		}
		return ids
	}()); got != "[t1]" {
		t.Errorf("auto targets = %s, want only the queued t1", got)
	}
	_ = root
}
