package factory

import "testing"

func dispatchCurrentTables(t *testing.T, s *BacklogStore) int {
	t.Helper()
	e, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.close() }()
	var n int
	if err := e.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='todo_dispatch_current'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func dispatchCurrentOf(t *testing.T, s *BacklogStore, card string) (run, owner string, ok bool) {
	t.Helper()
	record, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range record.Runtime.DispatchCurrent {
		if c.CardID == card {
			return c.RunID, c.OwnerLabel, true
		}
	}
	return "", "", false
}

// RefreshDispatchCurrentLockHeld only REFRESHES (card t1538, relay #3): a
// queue that never carried the record keeps its schema exactly as it was —
// the factory record's verbs leave the queue's schema alone (AC-FR-021).
func TestRefreshDispatchCurrentLeavesAQueueWithoutTheRecordUntouched(t *testing.T) {
	_, s, card := runtimeFixture(t)
	if got := dispatchCurrentTables(t, s); got != 0 {
		t.Fatalf("fixture queue already carries todo_dispatch_current (%d tables)", got)
	}
	if err := s.RefreshDispatchCurrentLockHeld(card, "run-1", "lane-1"); err != nil {
		t.Fatal(err)
	}
	if got := dispatchCurrentTables(t, s); got != 0 {
		t.Fatalf("a refresh created todo_dispatch_current in a queue that never had it")
	}
}

// An empty owner claims no owner — the record keeps the owner it holds for
// the SAME run and carries none onto a different run — and a card the
// dispatch hook never recorded gains no row.
func TestRefreshDispatchCurrentRefreshesOnlyAnExistingRecord(t *testing.T) {
	_, s, card := runtimeFixture(t)
	other, _, err := s.Add("other")
	if err != nil {
		t.Fatal(err)
	}
	// The dispatch hook's own write seeds the record for `card`.
	if err := s.recordRuntimeHook(TodoRuntimeRun{RunID: "run-1", ManifestJSON: "{}"}, &TodoRuntimeAssignment{
		RunID: "run-1", CardID: card, OwnerLabel: "worker-2", ReportedState: "picked", EventKind: "card.assigned", ProvenanceJSON: "{}",
	}, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if run, owner, ok := dispatchCurrentOf(t, s, card); !ok || run != "run-1" || owner != "lane-2" {
		t.Fatalf("seeded current = (%q, %q, %v), want (run-1, lane-2): the owner is stored canonical", run, owner, ok)
	}

	refresh := func(c, run, owner string) {
		t.Helper()
		if err := s.RefreshDispatchCurrentLockHeld(c, run, owner); err != nil {
			t.Fatal(err)
		}
	}
	refresh(card, "run-1", "")
	if run, owner, _ := dispatchCurrentOf(t, s, card); run != "run-1" || owner != "lane-2" {
		t.Fatalf("current = (%q, %q) after an owner-less refresh for the same run, want the owner kept", run, owner)
	}
	refresh(card, "run-2", "")
	if run, owner, _ := dispatchCurrentOf(t, s, card); run != "run-2" || owner != "" {
		t.Fatalf("current = (%q, %q) after an owner-less refresh for another run, want no owner carried over", run, owner)
	}
	refresh(card, "run-2", "worker-3")
	if run, owner, _ := dispatchCurrentOf(t, s, card); run != "run-2" || owner != "lane-3" {
		t.Fatalf("current = (%q, %q), want (run-2, lane-3)", run, owner)
	}

	refresh(other.ID, "run-9", "lane-9")
	if _, _, ok := dispatchCurrentOf(t, s, other.ID); ok {
		t.Fatal("a refresh inserted a record for a card the dispatch hook never recorded")
	}
	if err := s.RefreshDispatchCurrentLockHeld("", "run-2", "lane-3"); err == nil {
		t.Fatal("an empty card id was accepted")
	}
	if err := s.RefreshDispatchCurrentLockHeld(card, " ", "lane-3"); err == nil {
		t.Fatal("an empty run id was accepted")
	}
}
