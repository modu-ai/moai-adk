package auditreceipt

import (
	"errors"
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

// A ledger never written reads as the zero ledger, and an end recorded with
// no start on file is NOT counted into it: an end the ledger has no start for
// is not attributable to this background era (a foreground instance whose
// marker save failed resolves its end key here too — post-sync repair r5
// supplement 2) — counting it would fabricate a seal over a live era.
func TestInstanceLedgerAbsentReadsZeroAndEndWithoutStartIsNotCounted(t *testing.T) {
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
	if l.Ends != 0 || !l.EndedAt.IsZero() {
		t.Errorf("ledger = %+v, want untouched — an end without a recorded start must not aggregate into this ledger", l)
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
	// Widen the lock budget: 64 serialized updates must all land even while
	// the full suite loads the machine (an exhausted budget is a clean
	// give-up, and its own pin is TestLedgerLockTimesOutWhenHeld).
	prevWait := ledgerLockWait
	ledgerLockWait = 5 * time.Second
	t.Cleanup(func() { ledgerLockWait = prevWait })
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
	body := `{"pending_id":"idPend","ended_at":"` + endAt.Format(time.RFC3339Nano) + `","single_live":true}`
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
	if err := RecordInstanceStart(root, key2, t0.Add(500*time.Millisecond)); err != nil {
		t.Fatalf("second RecordInstanceStart key2: %v", err)
	}
	if err := os.WriteFile(ledgerPath(root, key2)+".uncertain", []byte("start-count uncertain\n"), 0o644); err != nil {
		t.Fatalf("mark uncertain: %v", err)
	}
	if err := os.WriteFile(ledgerPath(root, key2)+".end-pending-idPend2", []byte(`{"pending_id":"idPend2","ended_at":"`+endAt.Format(time.RFC3339Nano)+`","single_live":true}`), 0o644); err != nil {
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

// markInstanceEndPending writes the mark the next ledger operation replays.
func TestMarkInstanceEndPendingRoundTrip(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-pmark", AgentPlanAuditor)
	endAt := t0.Add(2 * time.Second)
	if err := markInstanceEndPending(root, key, endAt, true); err != nil {
		t.Fatalf("markInstanceEndPending: %v", err)
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

// Post-sync repair r3 (gate round 21, P1): the end-event boundary is a
// watermark — it must never regress. Two ends landing out of order (the 4s
// end applied before the 3s end; lock order is not event order) used to drag
// EndedAt back to 3s, and a successor could then reuse a receipt minted at
// 3.5s — after the true last end, before the regressed boundary.
func TestInstanceLedgerEndBoundaryIsMonotonic(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-mono", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil {
		t.Fatalf("RecordInstanceStart: %v", err)
	}
	// The later event is applied first; the earlier one lands afterwards.
	if err := RecordInstanceEnd(root, key, t0.Add(4*time.Second)); err != nil {
		t.Fatalf("first RecordInstanceEnd: %v", err)
	}
	if err := RecordInstanceEnd(root, key, t0.Add(3*time.Second)); err != nil {
		t.Fatalf("second RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if !l.EndedAt.Equal(t0.Add(4 * time.Second)) {
		t.Errorf("EndedAt = %v, want %v — the boundary must not regress", l.EndedAt, t0.Add(4*time.Second))
	}

	// A receipt minted between the two ends is predecessor-era: the seal holds.
	start := &StartMarker{AgentID: "a1", AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: t0}
	r35 := seedLedgerTestReceipt(t, root, t0.Add(3500*time.Millisecond))
	if ok, cause := CheckCitedReceiptsSince(root, start, l.EndedAt, []string{r35}); ok {
		t.Errorf("a receipt minted after the true last end but before the boundary was accepted — the boundary regressed")
	} else if cause != CauseReceiptReused {
		t.Errorf("cause = %q, want %q", cause, CauseReceiptReused)
	}
}

// Post-sync repair r3 (gate round 21, P2): the first start-marker write is a
// keep-earliest CAS — the first writer's anchor wins and LATER starts never
// overwrite it. The RED was observed through the pre-fix seam (8 concurrent
// plain WriteStartMarker calls kept the LATEST anchor: "REGRESSED anchor:
// StartedAt = ...340808, want ...280808"): the racy check-then-write then
// refused a legitimate receipt minted between the two starts with "receipt
// created before the auditor started". With claims racing, the deterministic
// property is first-writer-wins: an existing anchor survives any number of
// concurrent ensure calls.
func TestEnsureStartMarkerConcurrentKeepsExisting(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ensure-race", AgentPlanAuditor)
	anchor := time.Now().UTC()
	if err := WriteStartMarker(root, &StartMarker{AgentID: key, AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: anchor}); err != nil {
		t.Fatalf("seed anchor: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := StartMarker{AgentID: key, AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: anchor.Add(time.Duration(i+1) * 20 * time.Millisecond)}
			if err := EnsureStartMarker(root, &m); err != nil {
				t.Errorf("EnsureStartMarker %d: %v", i, err)
			}
		}()
	}
	wg.Wait()
	m, err := ReadStartMarker(root, key)
	if err != nil {
		t.Fatalf("marker missing: %v", err)
	}
	if !m.StartedAt.Equal(anchor) {
		t.Errorf("StartedAt = %v, want the untouched anchor %v — a later start overwrote it", m.StartedAt, anchor)
	}
}

// The CAS is serialized on the key's ledger lock: a marker ensure attempted
// while the lock is held past the wait budget gives up instead of writing
// outside the critical section.
func TestEnsureStartMarkerRespectsLedgerLock(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ensure-lock", AgentPlanAuditor)
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
	m := StartMarker{AgentID: key, AgentType: AgentPlanAuditor, TreeRoot: root}
	if err := EnsureStartMarker(root, &m); err == nil {
		t.Fatal("EnsureStartMarker wrote while the key's ledger lock was held — the CAS is not serialized")
	}
	if _, err := ReadStartMarker(root, key); err == nil {
		t.Error("a marker was written despite the held lock")
	}
}

// An existing anchor is never touched, whatever the new start claims.
func TestEnsureStartMarkerKeepsExistingAnchor(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ensure-keep", AgentPlanAuditor)
	if err := WriteStartMarker(root, &StartMarker{AgentID: key, AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: t0}); err != nil {
		t.Fatalf("seed marker: %v", err)
	}
	later := StartMarker{AgentID: key, AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: t0.Add(time.Minute)}
	if err := EnsureStartMarker(root, &later); err != nil {
		t.Fatalf("EnsureStartMarker: %v", err)
	}
	m, err := ReadStartMarker(root, key)
	if err != nil {
		t.Fatalf("ReadStartMarker: %v", err)
	}
	if !m.StartedAt.Equal(t0) {
		t.Errorf("StartedAt = %v, want the untouched earliest anchor %v", m.StartedAt, t0)
	}
}

// Post-sync repair r3 supplement (gate round 22): pending-file discovery must
// not interpret the project path as a glob pattern — a tree at
// .../project[1] made the discovery match nothing, so a dropped end was never
// recovered and predecessor receipts stayed unsealed. Discovery is a
// directory read with a literal filename prefix.
func TestInstanceLedgerPendingRecoveredUnderGlobMetacharPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project[1]")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	key := StartMarkerKey("", "sess-ledger-glob", AgentPlanAuditor)
	for i := 0; i < 2; i++ { // two recorded starts, two ends to account for
		if err := RecordInstanceStart(root, key, t0.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatalf("RecordInstanceStart %d: %v", i, err)
		}
	}
	pending := ledgerPath(root, key) + ".end-pending-idGlob"
	body := `{"pending_id":"idGlob","ended_at":"` + t0.Add(time.Second).Format(time.RFC3339Nano) + `"}`
	if err := os.WriteFile(pending, []byte(body), 0o644); err != nil {
		t.Fatalf("write pending mark: %v", err)
	}

	// The next operation discovers the pending mark despite the
	// metacharacter in the project path; its own end (2s, later than the
	// pending 1s) then seals the boundary as the watermark.
	if err := RecordInstanceEnd(root, key, t0.Add(2*time.Second)); err != nil {
		t.Fatalf("RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Ends != 2 {
		t.Errorf("Ends = %d, want 2 — the pending end was not discovered under a glob-metacharacter path", l.Ends)
	}
	if !l.EndedAt.Equal(t0.Add(2 * time.Second)) {
		t.Errorf("EndedAt = %v, want the watermark %v (the recovered end 1s, superseded by the later 2s)", l.EndedAt, t0.Add(2*time.Second))
	}
}

// Post-sync repair r5 (gate round 25, P2): two ends saved in reverse time
// order — the later end arrives while a sibling is still unaccounted for
// (ambiguous, discarded) and the earlier end then seals single-live. The
// boundary must still sit at the TRUE last terminal end: an ambiguous end
// that was discarded at its own arrival still happened, and once the era
// closes its timestamp bounds the seal — receipts minted between the two
// ends are predecessor-era.
func TestInstanceLedgerBoundarySealsAtLastEndWhenReordered(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-r5", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil {
		t.Fatalf("first RecordInstanceStart: %v", err)
	}
	if err := RecordInstanceStart(root, key, t0.Add(time.Second)); err != nil {
		t.Fatalf("second RecordInstanceStart: %v", err)
	}
	// The 3s end is applied first (two outstanding: ambiguous, discarded),
	// then the 2s end seals single-live.
	if err := RecordInstanceEnd(root, key, t0.Add(3*time.Second)); err != nil {
		t.Fatalf("first RecordInstanceEnd: %v", err)
	}
	if err := RecordInstanceEnd(root, key, t0.Add(2*time.Second)); err != nil {
		t.Fatalf("second RecordInstanceEnd: %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if !l.EndedAt.Equal(t0.Add(3 * time.Second)) {
		t.Errorf("EndedAt = %v, want the true last end %v — the boundary must not sit before the last terminal end of the era", l.EndedAt, t0.Add(3*time.Second))
	}

	// A receipt minted between the two ends is predecessor-era: refused.
	start := &StartMarker{AgentID: "a1", AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: t0}
	r25 := seedLedgerTestReceipt(t, root, t0.Add(2500*time.Millisecond))
	if ok, cause := CheckCitedReceiptsSince(root, start, l.EndedAt, []string{r25}); ok {
		t.Errorf("a receipt minted after the earlier end but before the true last end was accepted")
	} else if cause != CauseReceiptReused {
		t.Errorf("cause = %q, want %q", cause, CauseReceiptReused)
	}
}

// Post-sync repair r5 supplement 3 (gate round 33, P1): the pending replay
// must judge with the survivor state AT END TIME, not at replay time. A's
// single-live end write fails (its pending mark carries the judgment made
// then — A was the only instance outstanding), B's start is recorded before
// the replay, and the next ledger operation re-plays the end: re-deriving
// the judgment from the current count (B live, two outstanding) would
// misread it as ambiguous, consume the mark, and leave the boundary
// unsealed — the successor passes on A's receipt with no audit.
func TestInstanceLedgerPendingReplayKeepsEndTimeJudgment(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-r53", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil {
		t.Fatalf("RecordInstanceStart (A): %v", err)
	}
	// B is recorded BEFORE A's dropped end mark lands.
	if err := RecordInstanceStart(root, key, t0.Add(time.Second)); err != nil {
		t.Fatalf("RecordInstanceStart (B): %v", err)
	}
	// A's end happened at t0.5; its ledger save failed and the pending mark
	// carries the end-time judgment (single-live — A was alone).
	endAt := t0.Add(500 * time.Millisecond)
	pending := ledgerPath(root, key) + ".end-pending-idR53"
	body := `{"pending_id":"idR53","ended_at":"` + endAt.Format(time.RFC3339Nano) + `","single_live":true}`
	if err := os.WriteFile(pending, []byte(body), 0o644); err != nil {
		t.Fatalf("write pending mark: %v", err)
	}

	// The next ledger operation (C's start) replays the pending end: the
	// end-time judgment holds and A's era seals at t0.5, B live or not.
	if err := RecordInstanceStart(root, key, t0.Add(2*time.Second)); err != nil {
		t.Fatalf("RecordInstanceStart (replay carrier): %v", err)
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Ends != 1 {
		t.Errorf("Ends = %d, want 1 — the replayed end must be counted", l.Ends)
	}
	if !l.EndedAt.Equal(endAt) {
		t.Errorf("EndedAt = %v, want the end-time-sealed %v — the replay must keep the end-time judgment, not re-derive it as ambiguous", l.EndedAt, endAt)
	}

	// A receipt minted during A's lifetime is predecessor-era for any later
	// instance: refused.
	start := &StartMarker{AgentID: "a1", AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: t0}
	rA := seedLedgerTestReceipt(t, root, t0.Add(250*time.Millisecond))
	if ok, cause := CheckCitedReceiptsSince(root, start, l.EndedAt, []string{rA}); ok {
		t.Errorf("a receipt minted during A's lifetime was accepted after A's era sealed")
	} else if cause != CauseReceiptReused {
		t.Errorf("cause = %q, want %q", cause, CauseReceiptReused)
	}
}

// Post-sync repair r4 (gate round 25): a pending end mark that exists but
// cannot be read makes the boundary bookkeeping unknowable — a ledger update
// must give up (and the approval path must refuse), never apply a partial
// pending set or skip to the old boundary.
func TestInstanceLedgerUpdateGivesUpOnUnreadablePending(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-corrupt", AgentPlanAuditor)
	if err := RecordInstanceStart(root, key, t0); err != nil {
		t.Fatalf("RecordInstanceStart: %v", err)
	}
	pending := ledgerPath(root, key) + ".end-pending-idBad"
	if err := os.WriteFile(pending, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("write corrupt pending: %v", err)
	}

	if err := RecordInstanceEnd(root, key, t0.Add(time.Second)); err == nil {
		t.Fatal("the ledger update proceeded over an unreadable pending end")
	}
	l, err := ReadInstanceLedger(root, key)
	if err != nil {
		t.Fatalf("ReadInstanceLedger: %v", err)
	}
	if l.Ends != 0 || !l.EndedAt.IsZero() {
		t.Errorf("ledger = %+v, want untouched — no partial pending set may be applied", l)
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
	for i := 0; i < 2; i++ { // two recorded starts, two ends to account for
		if err := RecordInstanceStart(root, key, t0.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatalf("RecordInstanceStart %d: %v", i, err)
		}
	}
	pending := ledgerPath(root, key) + ".end-pending-idS"
	body := `{"pending_id":"idS","ended_at":"` + t0.Add(time.Second).Format(time.RFC3339Nano) + `"}`
	if err := os.WriteFile(pending, []byte(body), 0o644); err != nil {
		t.Fatalf("write pending: %v", err)
	}

	// Make the ledger save fail: swap the save seam. Platform-independent —
	// os.Chmod does not block file creation inside a directory on Windows, so
	// the previous permission-based injection never failed the save there
	// (post-sync repair r5).
	prevSave := ledgerSave
	ledgerSave = func(string, any) error { return errors.New("injected ledger save failure") }
	t.Cleanup(func() { ledgerSave = prevSave })
	if err := RecordInstanceEnd(root, key, t0.Add(2*time.Second)); err == nil {
		t.Fatal("the ledger save unexpectedly succeeded with the failing seam")
	}

	// The pending file survived the failed save...
	if _, err := os.Stat(pending); err != nil {
		t.Fatalf("the pending end was deleted before the save succeeded: %v", err)
	}

	// ...and the next operation recovers it.
	ledgerSave = prevSave
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

// PendingEndHold's three flavors: no pending holds nothing, a readable
// unapplied mark holds with ErrPendingEndUnresolved, a corrupted mark holds
// with ErrPendingEndUnreadable, and the applied mark (folded in by the next
// ledger operation) releases the hold.
func TestPendingEndHoldFlavors(t *testing.T) {
	root := t.TempDir()
	key := StartMarkerKey("", "sess-ledger-hold", AgentPlanAuditor)
	if hold := PendingEndHold(root, key); hold != nil {
		t.Fatalf("no pending end on file, got a hold: %v", hold)
	}

	if err := RecordInstanceStart(root, key, t0); err != nil {
		t.Fatalf("RecordInstanceStart: %v", err)
	}
	pending := ledgerPath(root, key) + ".end-pending-idHold"
	body := `{"pending_id":"idHold","ended_at":"` + t0.Add(time.Second).Format(time.RFC3339Nano) + `"}`
	if err := os.WriteFile(pending, []byte(body), 0o644); err != nil {
		t.Fatalf("write pending: %v", err)
	}
	hold := PendingEndHold(root, key)
	if !errors.Is(hold, ErrPendingEndUnresolved) {
		t.Fatalf("hold = %v, want ErrPendingEndUnresolved for a readable unapplied mark", hold)
	}

	// A corrupted mark holds with the unreadable flavor.
	if err := os.WriteFile(pending, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("corrupt pending: %v", err)
	}
	hold = PendingEndHold(root, key)
	if !errors.Is(hold, ErrPendingEndUnreadable) {
		t.Fatalf("hold = %v, want ErrPendingEndUnreadable for a corrupted mark", hold)
	}

	// The next ledger operation applies the (repaired) mark and consumes it:
	// the hold releases.
	if err := os.WriteFile(pending, []byte(body), 0o644); err != nil {
		t.Fatalf("repair pending: %v", err)
	}
	if err := RecordInstanceEnd(root, key, t0.Add(2*time.Second)); err != nil {
		t.Fatalf("RecordInstanceEnd (replay carrier): %v", err)
	}
	if hold := PendingEndHold(root, key); hold != nil {
		t.Fatalf("hold = %v, want nil after the pending end was applied", hold)
	}
}
