package auditreceipt

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// t0 anchors the ledger tests' clock.
var t0 = time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

// SPEC-RECEIPT-REUSE-001: a terminal end arriving while the ender was the only
// outstanding instance is single-live and advances the end-event boundary.
func TestInstanceLedgerSingleLiveEndAdvancesBoundary(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-1", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil {
		t.Fatalf("RecordInstanceStart: %v", err)
	}
	if err := RecordInstanceEnd(root, key, t0.Add(2*time.Second)); err != nil {
		t.Fatalf("RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Starts != 1 || l.Ends != 1 {
		t.Errorf("ledger = %+v, want 1 start and 1 end", l)
	}
	if !l.EndedAt.Equal(t0.Add(2 * time.Second)) {
		t.Errorf("EndedAt = %v, want the single-live end time %v", l.EndedAt, t0.Add(2*time.Second))
	}
}

// SPEC-RECEIPT-REUSE-001 AC-RR-009 sequence (ii): a terminal end arriving while
// MORE THAN ONE instance is outstanding is ambiguous and freezes the boundary;
// the surviving sibling's own single-live end then advances it.
func TestInstanceLedgerAmbiguousEndFreezesBoundary(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-2", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil { // A
		t.Fatalf("first RecordInstanceStart: %v", err)
	}
	if err := RecordInstanceStart(root, key, t0.Add(time.Second)); err != nil { // B
		t.Fatalf("second RecordInstanceStart: %v", err)
	}
	// B ends while two are outstanding: ambiguous, boundary frozen.
	if err := RecordInstanceEnd(root, key, t0.Add(3*time.Second)); err != nil {
		t.Fatalf("first RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if !l.EndedAt.IsZero() {
		t.Errorf("EndedAt = %v, want zero — an ambiguous end must not advance the boundary", l.EndedAt)
	}
	if l.Starts != 2 || l.Ends != 1 {
		t.Errorf("ledger = %+v, want 2 starts and 1 end", l)
	}
	// A's later end is single-live (2-1=1 outstanding): the boundary advances.
	if err := RecordInstanceEnd(root, key, t0.Add(5*time.Second)); err != nil {
		t.Fatalf("second RecordInstanceEnd: %v", err)
	}
	l, err = ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("second ReadInstanceLedger: %v", err)
	}
	if !l.EndedAt.Equal(t0.Add(5 * time.Second)) {
		t.Errorf("EndedAt = %v, want the surviving instance's end %v", l.EndedAt, t0.Add(5*time.Second))
	}
}

// A ledger never written reads as the zero ledger, and an end recorded with no
// start on file still seals forward (a later instance's own receipts are
// minted after its start, which is after the end).
func TestInstanceLedgerAbsentReadsZeroAndEndSealsForward(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-3", AgentSyncAuditor)
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger on absent ledger: %v", err)
	}
	if l.Starts != 0 || l.Ends != 0 || !l.EndedAt.IsZero() {
		t.Errorf("absent ledger = %+v, want the zero ledger", l)
	}
	if err := RecordInstanceEnd(root, key, t0.Add(time.Second)); err != nil {
		t.Fatalf("RecordInstanceEnd with no recorded start: %v", err)
	}
	l, err = ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Ends != 1 || !l.EndedAt.Equal(t0.Add(time.Second)) {
		t.Errorf("ledger = %+v, want 1 end at the end time", l)
	}
}

// SPEC-RECEIPT-REUSE-001 AC-RR-001/009: the end-event boundary refuses a
// receipt minted before a single-live predecessor end with the reuse cause,
// accepts one minted after it, and a zero boundary leaves the original check.
func TestCheckCitedReceiptsSinceEndBoundary(t *testing.T) {
	root := t.TempDir()
	start := &StartMarker{AgentID: "a1", AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: t0}
	before := seedLedgerTestReceipt(t, root, t0.Add(time.Second))  // minted t1, before the end t2
	after := seedLedgerTestReceipt(t, root, t0.Add(3*time.Second)) // minted t3, after the end t2
	end := t0.Add(2 * time.Second)

	if ok, cause := CheckCitedReceiptsSince(root, start, end, []string{before}); ok {
		t.Errorf("receipt minted before the boundary was accepted, want refusal with %q", CauseReceiptReused)
	} else if cause != CauseReceiptReused {
		t.Errorf("cause = %q, want %q", cause, CauseReceiptReused)
	}
	if ok, cause := CheckCitedReceiptsSince(root, start, end, []string{after}); !ok {
		t.Errorf("receipt minted after the boundary refused with %q, want acceptance", cause)
	}
	// A zero boundary is the original check: the before-era receipt passes the
	// start fence alone.
	if ok, cause := CheckCitedReceiptsSince(root, start, time.Time{}, []string{before}); !ok {
		t.Errorf("zero boundary refused a receipt after the start: %q", cause)
	}
	// The original entry point still reads as the zero boundary.
	if ok, cause := CheckCitedReceipts(root, start, []string{before}); !ok {
		t.Errorf("CheckCitedReceipts (no boundary) refused a receipt after the start: %q", cause)
	}
}

func seedLedgerTestReceipt(t *testing.T, root string, createdAt time.Time) string {
	t.Helper()
	id, err := WriteReceipt(root, &Receipt{Tool: ToolCodexAudit, TreeRoot: root, CreatedAt: createdAt})
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	return id
}

// Review repair (P1, card t1562 mid-flight review): concurrent starts and ends
// of one session are read-modify-writes on the SAME ledger file — without the
// lock, 32 concurrent starts survived as one. The lockfile serializes them.
func TestInstanceLedgerConcurrentCountsSurvive(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-conc", AgentPlanAuditor)
	const n = 32
	var wg sync.WaitGroup
	errs := make(chan error, 2*n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := RecordInstanceStart(root, key, t0); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := RecordInstanceEnd(root, key, t0.Add(time.Duration(i)*time.Millisecond)); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent ledger update failed: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Starts != n || l.Ends != n {
		t.Errorf("ledger = %d starts / %d ends, want %d / %d — a count was lost to a lost interleaving", l.Starts, l.Ends, n, n)
	}
	if l.EndedAt.IsZero() {
		t.Errorf("EndedAt zero after %d ends — the last single-live end must seal the era", n)
	}
}

// Post-sync repair (P2-1, card t1562 mid-flight review): a holder that stalls
// past the stale age can have its lock broken and taken by a successor. The
// stalled holder's eventual release must NOT delete the successor's lock —
// release is token-aware: it removes the lockfile only when the file still
// carries this holder's own token.
func TestLedgerLockReleaseRespectsSuccessorToken(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-own", AgentPlanAuditor)
	p := ledgerPath(root, key)

	release1, err := lockLedger(p)
	if err != nil {
		t.Fatalf("first lockLedger: %v", err)
	}
	// The first holder stalls past the stale age; a successor breaks the lock
	// and acquires its own.
	if err := os.Remove(p + ledgerLockSuffix); err != nil {
		t.Fatalf("successor break: %v", err)
	}
	release2, err := lockLedger(p)
	if err != nil {
		t.Fatalf("successor lockLedger: %v", err)
	}

	release1() // the stalled holder finally finishes
	if _, err := os.Stat(p + ledgerLockSuffix); err != nil {
		t.Errorf("the stalled holder's release deleted the successor's lock: %v", err)
	}

	release2() // the successor's own release removes its own lock
	if _, err := os.Stat(p + ledgerLockSuffix); !os.IsNotExist(err) {
		t.Errorf("the successor's release left its own lock behind: %v", err)
	}
}

// Post-sync repair (P2-2, card t1562 mid-flight review): a start whose ledger
// write was dropped leaves the outstanding count short by one — every end
// computed from it is unreliable, so the end path must FREEZE the boundary
// instead of advancing it on a count known to be incomplete. The uncertainty
// flag file is the durable trace of the dropped start; refusing a live
// instance's own receipt as predecessor-era (AC-RR-002) is the wrong error to
// make here.
func TestInstanceLedgerUncertainStartFreezesBoundary(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-unc", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil { // A is recorded
		t.Fatalf("RecordInstanceStart: %v", err)
	}
	// B's start-record write is dropped: the uncertainty is marked duratively.
	uncertain := ledgerPath(root, key) + ".uncertain"
	if err := os.WriteFile(uncertain, []byte("start-count uncertain\n"), 0o644); err != nil {
		t.Fatalf("mark uncertain: %v", err)
	}
	// A ends while the uncounted B is still live: the count reads 1 — do not
	// trust it.
	if err := RecordInstanceEnd(root, key, t0.Add(time.Second)); err != nil {
		t.Fatalf("RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if !l.EndedAt.IsZero() {
		t.Errorf("EndedAt = %v, want frozen — an end on a count known to be incomplete must not advance the boundary", l.EndedAt)
	}
	if l.Ends != 1 {
		t.Errorf("Ends = %d, want 1 — the end itself still counts", l.Ends)
	}
}

// Post-sync repair (P2-2 mirror, card t1562 gate round 19): an end whose
// ledger write was dropped leaves a pending mark, and the NEXT ledger
// operation replays it under the lock — the boundary seals at the pending end
// time. Unlike a dropped START (permanent freeze), a dropped END is
// recoverable: the end's own timestamp is known, only its write was lost.
func TestInstanceLedgerPendingEndReplaysUnderLock(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-pend", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil {
		t.Fatalf("RecordInstanceStart: %v", err)
	}
	// A's end write was dropped: the pending mark carries the entry id and the
	// end time.
	endAt := t0.Add(2 * time.Second)
	pending := ledgerPath(root, key) + ".end-pending-idPend"
	body := `{"pending_id":"idPend","ended_at":"` + endAt.Format(time.RFC3339Nano) + `"}`
	if err := os.WriteFile(pending, []byte(body), 0o644); err != nil {
		t.Fatalf("write pending mark: %v", err)
	}

	// The next operation — here B's start — replays the pending end first.
	if err := RecordInstanceStart(root, key, t0.Add(3*time.Second)); err != nil {
		t.Fatalf("RecordInstanceStart (replay carrier): %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if !l.EndedAt.Equal(endAt) {
		t.Errorf("EndedAt = %v, want the replayed pending end %v", l.EndedAt, endAt)
	}
	if l.Ends != 1 || l.Starts != 2 {
		t.Errorf("ledger = %d starts / %d ends, want 2 / 1 — the pending end must be counted", l.Starts, l.Ends)
	}
	if _, err := os.Stat(pending); !os.IsNotExist(err) {
		t.Errorf("the replayed pending mark was not consumed: %v", err)
	}

	// The uncertainty freeze outranks the replay: a pending end on a key whose
	// start count is known incomplete counts but never advances.
	key2 := StartMarkerKey("", "sess-ledger-pend-unc", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key2, t0); err != nil {
		t.Fatalf("RecordInstanceStart key2: %v", err)
	}
	if err := os.WriteFile(ledgerPath(root, key2)+".uncertain", []byte("start-count uncertain\n"), 0o644); err != nil {
		t.Fatalf("mark uncertain: %v", err)
	}
	if err := os.WriteFile(ledgerPath(root, key2)+".end-pending-idPend2", []byte(`{"pending_id":"idPend2","ended_at":"`+endAt.Format(time.RFC3339Nano)+`"}`), 0o644); err != nil {
		t.Fatalf("write pending mark key2: %v", err)
	}
	if err := RecordInstanceEnd(root, key2, t0.Add(4*time.Second)); err != nil {
		t.Fatalf("RecordInstanceEnd key2: %v", err)
	}
	l2, err := ReadInstanceLedger(root, key2)
	if err != nil {
		t.Fatalf("ReadInstanceLedger key2: %v", err)
	}
	if !l2.EndedAt.IsZero() {
		t.Errorf("key2 EndedAt = %v, want frozen — the uncertainty freeze outranks the replay", l2.EndedAt)
	}
	if l2.Ends != 2 {
		t.Errorf("key2 Ends = %d, want 2 — both the replayed and the current end count", l2.Ends)
	}
}

// MarkInstanceStartUncertain is idempotent: a second mark on an already
// uncertain key is not an error.
func TestMarkInstanceStartUncertainIdempotent(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-mark", AgentPlanAuditor)
	if err := MarkInstanceStartUncertain(root, key); err != nil {
		t.Fatalf("MarkInstanceStartUncertain: %v", err)
	}
	if err := MarkInstanceStartUncertain(root, key); err != nil {
		t.Fatalf("second MarkInstanceStartUncertain: %v", err)
	}
	if !startCountUncertain(root, key) {
		t.Errorf("the uncertainty mark is not visible to startCountUncertain")
	}
}

// MarkInstanceEndPending writes the mark the next ledger operation replays.
func TestMarkInstanceEndPendingRoundTrip(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-pmark", AgentPlanAuditor)
	endAt := t0.Add(2 * time.Second)
	if err := MarkInstanceEndPending(root, key, endAt); err != nil {
		t.Fatalf("MarkInstanceEndPending: %v", err)
	}
	marks, err := readEndPendings(root, key)
	if err != nil {
		t.Fatalf("readEndPendings: %v", err)
	}
	if len(marks) != 1 || !marks[0].pending.EndedAt.Equal(endAt) || marks[0].pending.PendingID == "" {
		t.Errorf("readEndPendings = %+v, want exactly the marked end %v under a fresh id", marks, endAt)
	}
}

// A ledger lock held past the wait budget makes lockLedger give up with an
// error rather than wait forever — the hook time budget bounds the wait.
func TestLedgerLockTimesOutWhenHeld(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-timeout", AgentPlanAuditor)
	p := ledgerPath(root, key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	lf, err := os.OpenFile(p+ledgerLockSuffix, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("hold the lock: %v", err)
	}
	defer func() {
		_ = lf.Close()
		_ = os.Remove(p + ledgerLockSuffix)
	}()
	if _, err := lockLedger(p); err == nil {
		t.Fatal("lockLedger acquired a lock held by another holder")
	}
}

// Post-sync supplement (gate round 20, P1): two same-role auditors can BOTH
// fail their end-record under lock contention. A single .end-pending file
// lets the second writer overwrite the first, so recovery loses an end and
// the boundary never seals. Each failed writer therefore gets its own
// per-instance pending file, and recovery applies ALL of them.
func TestInstanceLedgerAllPendingEndsRecovered(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-multi", AgentPlanAuditor)
	for i := 0; i < 3; i++ { // three recorded starts: A, B, C
		if err := RecordInstanceStart(root, key, t0.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatalf("RecordInstanceStart %d: %v", i, err)
		}
	}
	// A and B both fail their end-record: each leaves its own pending file.
	pendingA := ledgerPath(root, key) + ".end-pending-idA"
	pendingB := ledgerPath(root, key) + ".end-pending-idB"
	body := func(id string, at time.Time) string {
		return `{"pending_id":"` + id + `","ended_at":"` + at.Format(time.RFC3339Nano) + `"}`
	}
	if err := os.WriteFile(pendingA, []byte(body("idA", t0.Add(4*time.Second))), 0o644); err != nil {
		t.Fatalf("write pending A: %v", err)
	}
	if err := os.WriteFile(pendingB, []byte(body("idB", t0.Add(5*time.Second))), 0o644); err != nil {
		t.Fatalf("write pending B: %v", err)
	}

	// C's own end (the next ledger op) replays both pendings first, then
	// counts itself: all three ends land, and C — the last single-live end —
	// seals the boundary at its own time.
	if err := RecordInstanceEnd(root, key, t0.Add(6*time.Second)); err != nil {
		t.Fatalf("RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Ends != 3 {
		t.Errorf("Ends = %d, want 3 — a second failed writer must not overwrite the first's pending end", l.Ends)
	}
	if !l.EndedAt.Equal(t0.Add(6 * time.Second)) {
		t.Errorf("EndedAt = %v, want the last single-live end %v", l.EndedAt, t0.Add(6*time.Second))
	}
	if _, err := os.Stat(pendingA); !os.IsNotExist(err) {
		t.Errorf("pending A was not consumed: %v", err)
	}
	if _, err := os.Stat(pendingB); !os.IsNotExist(err) {
		t.Errorf("pending B was not consumed: %v", err)
	}
}

// Post-sync supplement (gate round 20, P2): the pending end used to be
// deleted BEFORE the ledger save, so a failed save lost the end from both
// sides. The pending file must survive a failed save and be recovered by the
// next operation.
func TestInstanceLedgerPendingSurvivesFailedSave(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-savefail", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil {
		t.Fatalf("RecordInstanceStart: %v", err)
	}
	pending := ledgerPath(root, key) + ".end-pending-idS"
	body := `{"pending_id":"idS","ended_at":"` + t0.Add(time.Second).Format(time.RFC3339Nano) + `"}`
	if err := os.WriteFile(pending, []byte(body), 0o644); err != nil {
		t.Fatalf("write pending: %v", err)
	}

	// Make the ledger save fail: the ledgers directory loses write permission,
	// so writeJSON's temp file cannot be created.
	dir := filepath.Dir(ledgerPath(root, key))
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("chmod ledgers dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	if err := RecordInstanceEnd(root, key, t0.Add(2*time.Second)); err == nil {
		t.Fatal("the ledger save unexpectedly succeeded in a read-only directory")
	}

	// The pending file survived the failed save...
	if _, err := os.Stat(pending); err != nil {
		t.Fatalf("the pending end was deleted before the save succeeded: %v", err)
	}

	// ...and the next operation recovers it.
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatalf("restore ledgers dir: %v", err)
	}
	if err := RecordInstanceEnd(root, key, t0.Add(3*time.Second)); err != nil {
		t.Fatalf("retrying RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Ends != 2 {
		t.Errorf("Ends = %d, want 2 — the recovered pending end plus the retry's own", l.Ends)
	}
	if _, err := os.Stat(pending); !os.IsNotExist(err) {
		t.Errorf("the recovered pending mark was not consumed: %v", err)
	}
}

// The replay is idempotent by pending-entry id: an entry already folded into
// a ledger save whose remove was lost (crash between save and remove) is
// skipped, not counted twice.
func TestInstanceLedgerPendingEndIdempotentByEntryID(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-idem", AgentPlanAuditor)
	if err := os.MkdirAll(filepath.Dir(ledgerPath(root, key)), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// A ledger whose save already folded in pending id "p1" (starts 2, ends 1),
	// with the pending file still on disk because its remove was lost.
	ledgerBody := `{"key":"` + key + `","starts":2,"ends":1,"applied_ends":["p1"],"updated_at":"` + t0.Format(time.RFC3339Nano) + `"}`
	if err := os.WriteFile(ledgerPath(root, key), []byte(ledgerBody), 0o644); err != nil {
		t.Fatalf("write ledger: %v", err)
	}
	pending := ledgerPath(root, key) + ".end-pending-p1"
	body := `{"pending_id":"p1","ended_at":"` + t0.Add(time.Second).Format(time.RFC3339Nano) + `"}`
	if err := os.WriteFile(pending, []byte(body), 0o644); err != nil {
		t.Fatalf("write pending: %v", err)
	}

	if err := RecordInstanceEnd(root, key, t0.Add(2*time.Second)); err != nil {
		t.Fatalf("RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Ends != 2 {
		t.Errorf("Ends = %d, want 2 — an already-applied pending entry must not be counted twice", l.Ends)
	}
	if _, err := os.Stat(pending); !os.IsNotExist(err) {
		t.Errorf("the already-applied pending mark was not consumed: %v", err)
	}
}
