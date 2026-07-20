package preference

// M4 crash-consistency + ScanDue/MarkScanned repro tests (SPEC-CLIFIX-CONCURRENCY-001
// REQ-CONC-001-005 / AC-CONC-001-005).
//
// This file references ONLY pre-M4 symbols (DecayScan, NewFileStore, ScanDue,
// MarkScanned, loadRecall, writeArchivalEntry, writeRecall) so it compiles
// against the pre-fix commit c31db9e2b. TestDecayCrash_* is a runtime RED
// against pre-M4 (the reconcile step does not exist, so a stale recall duplicate
// survives the scan). TestScanDueRace_* exercises concurrent MarkScanned under
// the race detector.

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// validTestEntry builds a minimal valid Entry for M4 tests. The caller may
// override fields (e.g. LastUsed to control age).
func validTestEntry(domain, key string) Entry {
	return Entry{
		Fact:           "test fact",
		SourceCitation: "m4_test.go",
		ValidTime:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		LastUsed:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Scope:          ScopeTransient,
		Domain:         domain,
		DecisionKey:    key,
		Confidence:     ConfidenceObserved,
		Weight:         1.0,
	}
}

// TestDecayCrash_ReconcilesDuplicateAfterInterruptedScan simulates the
// post-crash partial state (an entry landed in BOTH archival and recall because
// a prior DecayScan wrote archival then crashed before the recall write-back)
// and asserts the next DecayScan reconciles it away.
//
// To isolate the reconcile mechanism from the normal decay path, the seeded
// entry is NON-expired (age 0 at scan time): the normal decay logic would KEEP
// it in recall, so only the reconcile step can remove the stale recall copy.
//
// RED against pre-M4 (c31db9e2b): DecayScan has no reconcile step, so the stale
// recall duplicate survives → loadRecall returns [E] → the assertion fails.
// GREEN after M4: the reconcile step drops the recall copy → loadRecall [].
func TestDecayCrash_ReconcilesDuplicateAfterInterruptedScan(t *testing.T) {
	memDir := t.TempDir()
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	fs := store.(*fileStore)

	const domain, key = "test-domain", "crash-key"
	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	entry := validTestEntry(domain, key)
	entry.LastUsed = now // age 0 → non-expired; normal decay keeps it

	// Simulate the interrupted-scan partial state: archival written, recall
	// NOT cleaned. Both tiers hold the entry.
	if err := fs.writeArchivalEntry(domain, key, entry); err != nil {
		t.Fatalf("seed archival: %v", err)
	}
	if err := fs.writeRecall([]Entry{entry}); err != nil {
		t.Fatalf("seed recall: %v", err)
	}

	// Sanity: both tiers hold the duplicate before the scan.
	if pre, _ := fs.loadRecall(); len(pre) != 1 {
		t.Fatalf("pre-scan recall seed: want 1 entry, got %d", len(pre))
	}

	// Run the scan. age 0 → the normal decay path keeps the entry in recall;
	// only a reconcile step (drop recall copy when archival has it) cleans it.
	if _, err := fs.DecayScan(now); err != nil {
		t.Fatalf("DecayScan: %v", err)
	}

	// Post-scan: the stale recall duplicate MUST be gone (archival is the
	// authoritative copy; recall was not cleaned by the crashed prior scan).
	post, _ := fs.loadRecall()
	for _, e := range post {
		if e.Domain == domain && e.DecisionKey == key {
			t.Fatalf("stale recall duplicate survived scan: entry %s/%s still in recall (post=%v); archival also holds it → on-disk duplicate persists", domain, key, post)
		}
	}

	// The archival copy survives (the entry is soft-deleted to archival).
	if _, ok, err := fs.getFromArchival(domain, key); err != nil || !ok {
		t.Fatalf("archival copy lost after scan: ok=%v err=%v (reconcile must NOT drop the archival copy, only the stale recall duplicate)", ok, err)
	}
}

// TestScanDueRace_ConcurrentMarkScannedStaysValid drives N concurrent
// MarkScanned writers + M concurrent ScanDue readers under the race detector
// and asserts the stamp file is always a single valid parseable timestamp
// (never truncated, never corrupted).
//
// Under -race this also verifies no Go memory data race exists in the
// ScanDue/MarkScanned path. The atomic-write fix (MarkScanned → atomicWrite)
// guarantees a concurrent reader never observes a half-written (truncated) file.
func TestScanDueRace_ConcurrentMarkScannedStaysValid(t *testing.T) {
	stateDir := t.TempDir()
	base := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)

	const writers = 12
	const readers = 8
	const iters = 60

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	// Writers: hammer MarkScanned with distinct timestamps.
	for w := 0; w < writers; w++ {
		w := w
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				ts := base.Add(time.Duration(w*iters+i) * time.Second)
				if err := MarkScanned(stateDir, ts); err != nil {
					t.Errorf("MarkScanned: %v", err)
					return
				}
			}
		}()
	}
	// Readers: continuously poll ScanDue; every read must be error-free and
	// the underlying stamp (when present) must parse.
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				due, err := ScanDue(stateDir, base.Add(365*24*time.Hour))
				if err != nil {
					t.Errorf("ScanDue errored: %v", err)
					return
				}
				_ = due // boolean is advisory; the invariant under test is "no error / no corruption"
			}
		}()
	}
	wg.Wait()

	// Final invariant: the stamp file is a single valid parseable line.
	// The ScanDue call itself verifies parseStampTimestamp succeeds (no
	// corruption); the boolean is advisory and intentionally unused here —
	// the raw-file check below is the load-bearing assertion.
	if _, err := ScanDue(stateDir, base.Add(365 * 24 * time.Hour)); err != nil {
		t.Fatalf("final ScanDue: %v", err)
	}
	// Also confirm the raw file content is a single non-empty line (no
	// truncation, no interleaved writes). parseStampTimestamp already
	// validated it via ScanDue above; this is a belt-and-suspenders check.
	stampPath := filepath.Join(stateDir, decayLastRunFileName)
	raw, rerr := os.ReadFile(stampPath)
	if rerr != nil {
		t.Fatalf("read stamp: %v", rerr)
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		t.Fatalf("stamp file is empty/truncated after concurrent writes")
	}
	if _, perr := time.Parse(time.RFC3339, strings.SplitN(trimmed, "\n", 2)[0]); perr != nil {
		t.Fatalf("stamp file content not a valid RFC3339 timestamp: %q (%v)", trimmed, perr)
	}
}
