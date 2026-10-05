// classification_sort_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M2: the
// queue sort key (AC-TCD-005 at the store layer) and the classification
// decider seam's shipped implementations (REQ-TCD-012).
package factory

import (
	"fmt"
	"testing"
)

// classifyStore builds a record whose items carry the named classifications
// in slice (insertion) order.
func classifyStore(t *testing.T, defs []struct {
	id    string
	prio  string
	block bool
}) *BacklogRecord {
	t.Helper()
	rec := &BacklogRecord{Version: backlogVersion}
	for _, d := range defs {
		prio, mode := d.prio, ClassModeSerial
		if prio == "" {
			rec.Items = append(rec.Items, BacklogItem{ID: d.id, State: BacklogStateQueued})
			continue
		}
		cls := CardClassification{Priority: prio, Blocked: d.block, Mode: mode, Decider: DeciderIdentityLLM}
		rec.Items = append(rec.Items, BacklogItem{ID: d.id, State: BacklogStateQueued, Classification: &cls})
	}
	return rec
}

// TestSortByClassificationKeyOrder — AC-TCD-005: adding A(normal), B(high),
// C(low), D(normal, blocked) in that order sorts the queue B, A, C, D —
// blocked sinks behind every non-blocked card, priority orders the rest, and
// insertion order is stable within a rank.
func TestSortByClassificationKeyOrder(t *testing.T) {
	high := CardClassification{Priority: ClassPriorityHigh, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}
	low := CardClassification{Priority: ClassPriorityLow, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}
	blockedNormal := CardClassification{Priority: ClassPriorityNormal, Blocked: true, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}
	rec := &BacklogRecord{Version: backlogVersion, Items: []BacklogItem{
		{ID: "tA", State: BacklogStateQueued}, // unclassified = normal / non-blocked
		{ID: "tB", State: BacklogStateQueued, Classification: &high},
		{ID: "tC", State: BacklogStateQueued, Classification: &low},
		{ID: "tD", State: BacklogStateQueued, Classification: &blockedNormal},
	}}
	rec.SortByClassification()
	got := make([]string, 0, len(rec.Items))
	for _, it := range rec.Items {
		got = append(got, it.ID)
	}
	want := []string{"tB", "tA", "tC", "tD"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("sorted order = %v, want %v", got, want)
	}
}

// TestSortByClassificationMutationMovesRank — AC-TCD-005's mutation: the
// order responds to the rank, not to the sort call itself. Unblocking D
// (normal) re-sorts it ahead of the low card C; with the rank unchanged,
// stability keeps it where it was.
func TestSortByClassificationMutationMovesRank(t *testing.T) {
	high := CardClassification{Priority: ClassPriorityHigh, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}
	low := CardClassification{Priority: ClassPriorityLow, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}
	rec := &BacklogRecord{Version: backlogVersion, Items: []BacklogItem{
		{ID: "tA", State: BacklogStateQueued},
		{ID: "tB", State: BacklogStateQueued, Classification: &high},
		{ID: "tC", State: BacklogStateQueued, Classification: &low},
		{ID: "tD", State: BacklogStateQueued, Classification: &CardClassification{Priority: ClassPriorityNormal, Blocked: true, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}},
	}}
	rec.SortByClassification()
	if got := rec.Items[0].ID; got != "tB" || rec.Items[3].ID != "tD" {
		t.Fatalf("blocked D did not sink behind C: %v", rec.Items)
	}
	// Mutate: D becomes non-blocked (same normal priority).
	*rec.Items[3].Classification = CardClassification{Priority: ClassPriorityNormal, Blocked: false, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}
	rec.SortByClassification()
	got := []string{rec.Items[0].ID, rec.Items[1].ID, rec.Items[2].ID, rec.Items[3].ID}
	if want := []string{"tB", "tA", "tD", "tC"}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("after unblocking D, order = %v, want %v (D re-sorted ahead of low C)", got, want)
	}
}

// TestSortByClassificationStableWithinRank — same rank keeps insertion order
// (the tie-break is the existing order, never the id or the text).
func TestSortByClassificationStableWithinRank(t *testing.T) {
	rec := classifyStore(t, []struct {
		id    string
		prio  string
		block bool
	}{
		{"t9", "", false}, {"t3", "", false}, {"t7", "", false},
	})
	rec.SortByClassification()
	for i, want := range []string{"t9", "t3", "t7"} {
		if rec.Items[i].ID != want {
			t.Errorf("rank position %d = %s, want %s (insertion order must be stable)", i, rec.Items[i].ID, want)
		}
	}
}

// TestSortByClassificationTouchesEveryState — the sort is the record's one
// canonical order: queued, picked, and hold cards all take their rank
// position. (HOLD-STATE interplay: sorting positions a held card, selection
// filters it — constraint C3.)
func TestSortByClassificationTouchesEveryState(t *testing.T) {
	high := CardClassification{Priority: ClassPriorityHigh, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}
	rec := &BacklogRecord{Version: backlogVersion, Items: []BacklogItem{
		{ID: "tHold", State: BacklogStateHold, Classification: &high},
		{ID: "tQueued", State: BacklogStateQueued},
	}}
	rec.SortByClassification()
	if rec.Items[0].ID != "tHold" {
		t.Errorf("high-priority held card did not take the front: %v", rec.Items)
	}
}

// TestQueuedPositionAfterSort — AC-TCD-006: the position add prints is the
// 1-based position in the SORTED queued order, not the append order.
func TestQueuedPositionAfterSort(t *testing.T) {
	high := CardClassification{Priority: ClassPriorityHigh, Mode: ClassModeSerial, Decider: DeciderIdentityLLM}
	rec := &BacklogRecord{Version: backlogVersion, Items: []BacklogItem{
		{ID: "t1", State: BacklogStateQueued},
		{ID: "t2", State: BacklogStateQueued, Classification: &high},
		{ID: "t3", State: BacklogStatePicked},
	}}
	rec.SortByClassification()
	if got := rec.QueuedPosition("t2"); got != 1 {
		t.Errorf("t2 queued position = %d, want 1 (high sorts first)", got)
	}
	if got := rec.QueuedPosition("t1"); got != 2 {
		t.Errorf("t1 queued position = %d, want 2", got)
	}
	if got := rec.QueuedPosition("t3"); got != 0 {
		t.Errorf("picked t3 queued position = %d, want 0 (not queued)", got)
	}
}

// TestDefaultCardDeciderIsDeterministic — REQ-TCD-012: the shipped default
// implementation always answers with the fail-safe defaults and never
// errors; it is a judgment (decider default), not a failure.
func TestDefaultCardDeciderIsDeterministic(t *testing.T) {
	for i := 0; i < 3; i++ {
		c, err := DefaultCardDecider{}.Classify("any text")
		if err != nil {
			t.Fatalf("default decider errored: %v", err)
		}
		want := DefaultCardClassification()
		if c.Priority != want.Priority || c.Blocked != want.Blocked || c.Mode != want.Mode || c.Decider != want.Decider {
			t.Errorf("default decider produced %+v, want the fail-safe defaults %+v", c, want)
		}
	}
}

// TestStaticCardDeciderCarriesSuppliedJudgment — the --classification-file
// transport (REQ-TCD-004): a validated supplied judgment travels the seam
// verbatim, decider identity included.
func TestStaticCardDeciderCarriesSuppliedJudgment(t *testing.T) {
	cls := CardClassification{Priority: ClassPriorityLow, Blocked: true, Mode: ClassModeParallelizable, Decider: DeciderIdentityLLM, Reason: "supplied"}
	got, err := StaticCardDecider{Class: cls}.Classify("text")
	if err != nil {
		t.Fatalf("static decider errored: %v", err)
	}
	if got != cls {
		t.Errorf("static decider produced %+v, want the supplied judgment %+v", got, cls)
	}
}
