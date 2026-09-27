package homestate

import (
	"sync"
	"testing"
	"time"
)

// An entry appended while a reconciliation rewrite is between its read and
// its rename must survive the rewrite. The hook starts the append at exactly
// that point and gives it up to appendWait to finish; without serialization
// the append lands in the file the rename is about to replace and is lost.
func TestFR_UnavailableLogAppendSurvivesRewrite(t *testing.T) {
	root := factorySandbox(t)
	if err := AppendRecordUnavailable(root, RecordUnavailableEntry{ID: "first", RunID: frRun, CardID: "lost", Lane: "worker-2", Error: "boom"}); err != nil {
		t.Fatal(err)
	}
	path, err := RecordUnavailablePath(root)
	if err != nil {
		t.Fatal(err)
	}
	const appendWait = 500 * time.Millisecond
	var wg sync.WaitGroup
	appendErr := make(chan error, 1)
	recordUnavailableRewriteHook = func() {
		recordUnavailableRewriteHook = nil
		done := make(chan struct{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			appendErr <- AppendRecordUnavailable(root, RecordUnavailableEntry{ID: "concurrent", RunID: frRun, CardID: "late", Lane: "worker-3", Error: "boom"})
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(appendWait):
		}
	}
	t.Cleanup(func() { recordUnavailableRewriteHook = nil })
	if err := markRecordUnavailableReconciled(path, map[string]bool{"first": true}); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	wg.Wait()
	if err := <-appendErr; err != nil {
		t.Fatalf("concurrent append: %v", err)
	}
	entries, _, err := readRecordUnavailableFile(path)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]RecordUnavailableEntry{}
	for _, e := range entries {
		seen[e.ID] = e
	}
	if !seen["first"].Reconciled {
		t.Fatalf("first entry not reconciled: %+v", entries)
	}
	late, ok := seen["concurrent"]
	if !ok {
		t.Fatalf("the entry appended during the rewrite was lost: %+v", entries)
	}
	if late.Reconciled {
		t.Fatalf("the concurrent entry was reconciled without a drift event: %+v", late)
	}
}

// Many appenders racing one reconciliation lose nothing (run under -race).
func TestFR_UnavailableLogConcurrentAppendsNoLoss(t *testing.T) {
	root := factorySandbox(t)
	if err := AppendRecordUnavailable(root, RecordUnavailableEntry{ID: "seed", RunID: frRun, CardID: "c0", Lane: "w", Error: "e"}); err != nil {
		t.Fatal(err)
	}
	path, _ := RecordUnavailablePath(root)
	const n = 20
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "a" + string(rune('a'+i))
			if err := AppendRecordUnavailable(root, RecordUnavailableEntry{ID: id, RunID: frRun, CardID: id, Lane: "w", Error: "e"}); err != nil {
				t.Error(err)
			}
		}(i)
		if i == n/2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := markRecordUnavailableReconciled(path, map[string]bool{"seed": true}); err != nil {
					t.Error(err)
				}
			}()
		}
	}
	wg.Wait()
	entries, _, err := readRecordUnavailableFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != n+1 {
		t.Fatalf("entries = %d, want %d (an append was lost)", len(entries), n+1)
	}
}
