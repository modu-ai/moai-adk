package homestate

// SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458) AC-FAL-015 at the record write:
// fixture (c) — the drift log's lock held for 1.5 s while one write is made
// through the lease claim's opt-in (WithBoundedReconcile) and one through the
// ordinary path — and fixture (e), the re-read under the lock. Plus the
// layering guard of AC-FAL-011: internal/homestate's non-test files have no
// dependency path to internal/factory.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// flcHold is the hold of fixture (c).
const flcHold = 1500 * time.Millisecond

// flcFixture opens a sandboxed factory record with one run, one registered
// worker and one unreconciled drift-log entry for the run, and returns the log
// path and the entry's id.
func flcFixture(t *testing.T) (*FactoryDB, string, string) {
	t.Helper()
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-p", Backend: "claude", LeadPID: 1, LeadProcessStart: "st"})
	frRegisterWorker(t, db, "lane-1")
	logPath := filepath.Join(filepath.Dir(db.Path), recordUnavailableFile)
	const id = "held-0"
	line, err := json.Marshal(RecordUnavailableEntry{ID: id, At: "2026-10-03T00:00:00Z", RunID: "run-p", CardID: "lost-1", Lane: "lane-2", Error: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, append(line, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return db, logPath, id
}

// flcUnreconciled counts the run's unreconciled entries in the log.
func flcUnreconciled(t *testing.T, path string) int {
	t.Helper()
	entries, _, err := readRecordUnavailableFile(path)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range entries {
		if !e.Reconciled && e.RunID == "run-p" {
			n++
		}
	}
	return n
}

// flcHoldLock takes the log's lock with the primitive the log uses and releases
// it after d; the returned function waits for the release.
func flcHoldLock(t *testing.T, path string, d time.Duration) (released func() time.Time) {
	t.Helper()
	unlock, err := lockRecordUnavailable(path)
	if err != nil {
		t.Fatal(err)
	}
	at := make(chan time.Time, 1)
	time.AfterFunc(d, func() {
		now := time.Now()
		unlock()
		at <- now
	})
	var got time.Time
	return func() time.Time {
		if got.IsZero() {
			got = <-at
		}
		return got
	}
}

// TestRecordWriteReconcileBoundedSkipsOnContention — AC-FAL-015 (c), clause (iv),
// the claim's half: with the log's lock held for 1.5 s, a record write made
// through the claim's opt-in returns within 500 ms and leaves the entry
// unreconciled, appending no record.drift event (the skip withholds both).
func TestRecordWriteReconcileBoundedSkipsOnContention(t *testing.T) {
	db, logPath, _ := flcFixture(t)
	released := flcHoldLock(t, logPath, flcHold)
	ctx := WithBoundedReconcile(context.Background())
	start := time.Now()
	_, err := db.RecordPicked(ctx, "run-p", "t9701", CardFields{}, "probe", time.Now())
	elapsed := time.Since(start)
	unrec, drift := flcUnreconciled(t, logPath), len(frEvents(t, db, "record.drift"))
	released()
	t.Logf("opt-in write under a held drift-log lock: elapsed=%s err=%v unreconciled-after=%d record.drift-events=%d", elapsed, err, unrec, drift)
	if err != nil {
		t.Fatalf("the opt-in write failed: %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("the opt-in write returned after %s with the drift-log lock held, want within 500ms (it must not wait for the log's lock)", elapsed)
	}
	if unrec != 1 || drift != 0 {
		t.Errorf("the opt-in write left unreconciled=%d (want 1) and record.drift-events=%d (want 0): a skipped reconciliation withholds both the events and the mark", unrec, drift)
	}
}

// TestRecordWriteReconcileDefaultStillWaits — AC-FAL-015 (c), clause (iv), the
// ordinary half: with the log's lock held for 1.5 s, a write made through the
// ordinary path (as `factory assign` makes it) waits until the lock is released
// and then reconciles the entry — one event, the entry marked. The skip is the
// claim's alone (REQ-FAL-014's last sentence).
func TestRecordWriteReconcileDefaultStillWaits(t *testing.T) {
	db, logPath, _ := flcFixture(t)
	released := flcHoldLock(t, logPath, flcHold)
	start := time.Now()
	_, err := db.RecordPicked(context.Background(), "run-p", "t9702", CardFields{}, "probe", time.Now())
	returned := time.Now()
	rel := released()
	unrec, drift := flcUnreconciled(t, logPath), len(frEvents(t, db, "record.drift"))
	t.Logf("ordinary write under a held drift-log lock: elapsed=%s returned-vs-release=%s err=%v unreconciled-after=%d record.drift-events=%d", returned.Sub(start), returned.Sub(rel), err, unrec, drift)
	if err != nil {
		t.Fatalf("the ordinary write failed: %v", err)
	}
	if returned.Before(rel) {
		t.Errorf("the ordinary write returned %s before the lock was released; it must wait for the held lock", rel.Sub(returned))
	}
	if unrec != 0 || drift != 1 {
		t.Errorf("the ordinary write left unreconciled=%d (want 0) and record.drift-events=%d (want 1)", unrec, drift)
	}
}

// TestRecordWriteReconcileBoundedRereadsUnderLock — AC-FAL-015 (e), clause (vi):
// the entry is unreconciled when the claim's write reads the log, and the test
// seam between that unlocked read and the claim's try for the lock marks it
// reconciled (the mark another writer's post-commit step makes) while the lock
// is free. The claim's write re-reads the log under the lock, finds the entry
// already reconciled, and appends NO record.drift event; the entry stays
// marked.
func TestRecordWriteReconcileBoundedRereadsUnderLock(t *testing.T) {
	db, logPath, id := flcFixture(t)
	called := false
	prev := recordUnavailableAfterReadHook
	recordUnavailableAfterReadHook = func() {
		called = true
		if err := markRecordUnavailableReconciled(logPath, map[string]bool{id: true}); err != nil {
			t.Errorf("the seam could not mark the entry: %v", err)
		}
	}
	t.Cleanup(func() { recordUnavailableAfterReadHook = prev })

	_, err := db.RecordPicked(WithBoundedReconcile(context.Background()), "run-p", "t9703", CardFields{}, "probe", time.Now())
	unrec, drift := flcUnreconciled(t, logPath), len(frEvents(t, db, "record.drift"))
	t.Logf("opt-in write with the entry marked between its read and its lock: seam-called=%v err=%v unreconciled-after=%d record.drift-events=%d", called, err, unrec, drift)
	if err != nil {
		t.Fatalf("the opt-in write failed: %v", err)
	}
	if !called {
		t.Errorf("the seam between the unlocked read and the claim's try for the lock was never called (the claim-scoped reconciliation is not wired)")
	}
	if drift != 0 {
		t.Errorf("the claim appended %d record.drift event(s) for an entry another writer had already marked, want 0 (the claim must re-read the log under the lock)", drift)
	}
	if unrec != 0 {
		t.Errorf("unreconciled=%d after the write, want 0 (the entry stays marked)", unrec)
	}
}

// TestHomestateDoesNotImportFactory — AC-FAL-011's layering guard:
// internal/homestate has no dependency path to internal/factory through its
// NON-TEST files. The guard lists `go list -deps` of the package without
// `-test`: a test file that imports internal/factory (temp_parity_test.go does)
// does not turn it red (ledger L16).
func TestHomestateDoesNotImportFactory(t *testing.T) {
	const pkg = "github.com/modu-ai/moai-adk/internal/homestate"
	const factoryPkg = "github.com/modu-ai/moai-adk/internal/factory"
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("the go tool is not on PATH: %v", err)
	}
	out, err := exec.Command(goBin, "list", "-deps", pkg).Output()
	if err != nil {
		t.Fatalf("go list -deps %s: %v", pkg, err)
	}
	deps := strings.Fields(string(out))
	// Positive control: the listing is real (it names the package itself and a
	// dependency the package certainly has), so a missing factory line is a
	// measurement and not an empty listing.
	hasSelf, hasSQL := false, false
	for _, d := range deps {
		switch d {
		case pkg:
			hasSelf = true
		case "database/sql":
			hasSQL = true
		case factoryPkg:
			t.Errorf("internal/homestate's non-test files depend on %s (a layering violation: the queue lock sits above the record)", factoryPkg)
		}
	}
	if !hasSelf || !hasSQL {
		t.Fatalf("the dependency listing is not credible (self=%v database/sql=%v, %d entries)", hasSelf, hasSQL, len(deps))
	}
}
