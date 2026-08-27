// backlog_concurrency_test.go — the lost-update and id-collision guards under
// contention (SPEC-TODO-SQLITE-001 AC-TOSQ-009, REQ-TOSQ-005/008; M2).
//
// Factory mode runs up to ten lanes against ONE queue. The storage swap must
// not weaken that: the outer advisory lock still serializes the whole
// read-modify-write, and the engine's UNIQUE(id) plus the single-transaction
// last_seq advance are the backstop underneath it. Both are exercised here by
// contention, not by inspection.
package kanban

import (
	"path/filepath"
	"sync"
	"testing"
)

// AC-TOSQ-009 / REQ-TOSQ-008: N concurrent adders over one seeded root yield N
// unique sequential ids and zero lost updates. Run with -race.
func TestConcurrencyStress(t *testing.T) {
	t.Parallel()
	const (
		writers       = 8
		addsPerWriter = 6
		wantTotal     = writers * addsPerWriter
	)
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))

	var mu sync.Mutex
	issued := make(map[string]int, wantTotal)
	failures := make([]error, 0)

	var start sync.WaitGroup
	start.Add(1)
	var done sync.WaitGroup
	for w := 0; w < writers; w++ {
		done.Add(1)
		go func(w int) {
			defer done.Done()
			start.Wait() // release every writer into the lock at once
			for i := 0; i < addsPerWriter; i++ {
				item, _, err := store.Add("card")
				mu.Lock()
				if err != nil {
					failures = append(failures, err)
				} else {
					issued[item.ID]++
				}
				mu.Unlock()
			}
		}(w)
	}
	start.Done()
	done.Wait()

	if len(failures) != 0 {
		t.Fatalf("%d/%d adds failed under contention; first: %v", len(failures), wantTotal, failures[0])
	}

	// Zero collisions: every issued id is distinct.
	for id, n := range issued {
		if n != 1 {
			t.Errorf("id %s issued %d times — the mark advanced outside the lock", id, n)
		}
	}
	if len(issued) != wantTotal {
		t.Fatalf("distinct ids = %d, want %d", len(issued), wantTotal)
	}

	// Zero lost updates: every issued card is actually in the queue. A lost
	// update is invisible in the id set — it shows up only here, as an id that
	// was handed out and then overwritten by a concurrent whole-record write.
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load after stress: %v", err)
	}
	if len(rec.Items) != wantTotal {
		t.Fatalf("stored items = %d, want %d — %d updates were lost",
			len(rec.Items), wantTotal, wantTotal-len(rec.Items))
	}
	stored := make(map[string]bool, len(rec.Items))
	for _, it := range rec.Items {
		stored[it.ID] = true
	}
	for id := range issued {
		if !stored[id] {
			t.Errorf("id %s was issued but is not in the queue — lost update", id)
		}
	}
	if rec.LastSeq != wantTotal {
		t.Errorf("last_seq = %d, want %d", rec.LastSeq, wantTotal)
	}
	t.Logf("AC-TOSQ-009: %d writers x %d adds = %d distinct ids, %d stored items, last_seq %d, 0 collisions, 0 lost updates",
		writers, addsPerWriter, len(issued), len(rec.Items), rec.LastSeq)
}

// REQ-TOSQ-005: id integrity is enforced at the STORAGE layer, not only by the
// issuing code. A mutation that mints a duplicate id must abort the whole
// transaction and leave the prior state intact — fired here by handing the
// store a callback that deliberately duplicates an existing id, which is the
// shape a mis-ported issuer or a hand-edited import would take.
func TestDuplicateIDRejectedByStorage(t *testing.T) {
	t.Parallel()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if _, _, err := store.Add("original"); err != nil {
		t.Fatalf("seed Add: %v", err)
	}

	err := store.Mutate(func(rec *BacklogRecord) error {
		dup := rec.Items[0]
		dup.Text = "duplicate id smuggled in"
		rec.Items = append(rec.Items, dup)
		return nil
	})
	if err == nil {
		t.Fatal("Mutate(duplicate id) err = nil, want a storage-layer rejection")
	}
	if !IsBacklogIDConflict(err) {
		t.Errorf("err = %v, want it to satisfy IsBacklogIDConflict", err)
	}

	rec, loadErr := store.Load()
	if loadErr != nil {
		t.Fatalf("Load after rejected mutation: %v", loadErr)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("items = %d, want 1 — the aborted transaction must leave prior state intact", len(rec.Items))
	}
	if rec.Items[0].Text != "original" {
		t.Errorf("surviving item text = %q, want %q", rec.Items[0].Text, "original")
	}
}

// REQ-TOSQ-004: a `todo move`-style reorder round-trips. This is the case the
// design's "seq mirrors t<N>" note would have broken — after a move, array
// order and id order differ, and array order is the contract.
func TestReorderedItemsRoundTrip(t *testing.T) {
	t.Parallel()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	for _, text := range []string{"first", "second", "third"} {
		if _, _, err := store.Add(text); err != nil {
			t.Fatalf("Add(%s): %v", text, err)
		}
	}

	// Move the last card to the front, exactly as `todo move` rewrites the slice.
	if err := store.Mutate(func(rec *BacklogRecord) error {
		last := rec.Items[len(rec.Items)-1]
		rec.Items = append([]BacklogItem{last}, rec.Items[:len(rec.Items)-1]...)
		return nil
	}); err != nil {
		t.Fatalf("reorder Mutate: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load after reorder: %v", err)
	}
	got := []string{rec.Items[0].ID, rec.Items[1].ID, rec.Items[2].ID}
	want := []string{"t3", "t1", "t2"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order after reorder = %v, want %v (array position is the contract)", got, want)
		}
	}
}
