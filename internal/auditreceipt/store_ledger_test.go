package auditreceipt

import (
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
