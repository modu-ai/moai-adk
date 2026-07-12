package routing

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fixedClock is a deterministic clock for age-guarded sweep tests (AP-5: no
// wall-clock sleeps or timing assertions).
func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func writeForeignPending(t *testing.T, dir, sessionID string, createdAt time.Time) {
	t.Helper()
	p := PendingRow{
		SchemaVersion:     SchemaVersion,
		CreatedAt:         createdAt,
		TS:                createdAt.Format(time.RFC3339),
		SessionID:         sessionID,
		ModelClass:        "unknown",
		MatchedSubcommand: "run",
		RequestDigest:     "sha256:ffffffffffff",
		EvidenceRefs:      []EvidenceRef{},
		Delegations:       []Delegation{},
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, pendingPrefix+sessionKey(sessionID)+pendingSuffix)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readLedger(t *testing.T, dir string) []Row {
	t.Helper()
	rows, skipped, err := NewReader(filepath.Join(dir, LedgerFileName)).Read(Filter{})
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if skipped != 0 {
		t.Fatalf("unexpected %d malformed ledger lines", skipped)
	}
	return rows
}

func pendingExists(dir, sessionID string) bool {
	_, err := os.Stat(filepath.Join(dir, pendingPrefix+sessionKey(sessionID)+pendingSuffix))
	return err == nil
}

func recordRow(sessionID, sub string) PendingRow {
	return PendingRow{
		SessionID:         sessionID,
		MatchedSubcommand: sub,
		RequestDigest:     RequestDigest(sub + " request"),
		RequestClass:      ClassifyRequest(sub + " request"),
	}
}

// TestReroute asserts same-session re-record finalizes the prior pending row as
// reroute before the new pending row is created (REQ-HEV-010, AC-HEV-014).
func TestReroute(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)

	if err := s.Record(recordRow("sess-A", "plan")); err != nil {
		t.Fatal(err)
	}
	if err := s.Record(recordRow("sess-A", "run")); err != nil {
		t.Fatal(err)
	}

	rows := readLedger(t, dir)
	if len(rows) != 1 {
		t.Fatalf("got %d ledger rows, want 1 (the rerouted plan row)", len(rows))
	}
	if rows[0].Outcome != OutcomeReroute || rows[0].MatchedSubcommand != "plan" {
		t.Fatalf("rerouted row = %+v; want outcome=reroute subcommand=plan", rows[0])
	}
	if !pendingExists(dir, "sess-A") {
		t.Fatal("new pending row for sess-A should exist after re-record")
	}
}

// TestReroute_AgeIndependent: even an OLD same-session row reroutes, never
// aborts (REQ-HEV-010 precedence over REQ-HEV-014).
func TestReroute_AgeIndependent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	s := NewStore(dir).WithClock(fixedClock(now))

	// A 30h-old same-session pending row.
	writeForeignPending(t, dir, "self", now.Add(-30*time.Hour))
	if err := s.Record(recordRow("self", "run")); err != nil {
		t.Fatal(err)
	}
	rows := readLedger(t, dir)
	if len(rows) != 1 || rows[0].Outcome != OutcomeReroute {
		t.Fatalf("old same-session row must reroute, got %+v", rows)
	}
}

// TestStaleSweepAbort covers the age-guarded + liveness-guarded sweep matrix
// (REQ-HEV-014, AC-HEV-017 a/b/c).
func TestStaleSweepAbort(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)

	t.Run("a: foreign row older than 24h, not live -> swept abort", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		s := NewStore(dir).WithClock(fixedClock(now))
		writeForeignPending(t, dir, "old-dead", now.Add(-25*time.Hour))

		if err := s.Record(recordRow("cur", "run")); err != nil {
			t.Fatal(err)
		}
		rows := readLedger(t, dir)
		if len(rows) != 1 || rows[0].Outcome != OutcomeAbort || rows[0].SessionID != "old-dead" {
			t.Fatalf("expected 1 abort row for old-dead, got %+v", rows)
		}
		if pendingExists(dir, "old-dead") {
			t.Fatal("swept foreign pending file should be removed")
		}
		if !pendingExists(dir, "cur") {
			t.Fatal("new pending for cur should exist")
		}
	})

	t.Run("b: foreign row younger than 24h -> untouched", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		s := NewStore(dir).WithClock(fixedClock(now))
		writeForeignPending(t, dir, "young", now.Add(-1*time.Hour))

		if err := s.Record(recordRow("cur", "run")); err != nil {
			t.Fatal(err)
		}
		if len(readLedger(t, dir)) != 0 {
			t.Fatal("young foreign row must NOT be aborted (live parallel session protection)")
		}
		if !pendingExists(dir, "young") {
			t.Fatal("young foreign pending file must survive")
		}
	})

	t.Run("c: foreign row older than 24h but session live -> untouched", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		s := NewStore(dir).WithClock(fixedClock(now))
		writeForeignPending(t, dir, "old-live", now.Add(-25*time.Hour))
		// Best-effort liveness registry lists old-live as a live session.
		if err := os.WriteFile(filepath.Join(dir, activeSessionsFile),
			[]byte(`[{"session_id":"old-live","pid":123}]`), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := s.Record(recordRow("cur", "run")); err != nil {
			t.Fatal(err)
		}
		if len(readLedger(t, dir)) != 0 {
			t.Fatal("live-listed foreign row must NOT be aborted regardless of age")
		}
		if !pendingExists(dir, "old-live") {
			t.Fatal("live foreign pending file must survive")
		}
	})
}

// TestFinalize_SelfGate_NoPendingNoOp: Stop finalize with no pending row is a
// pure no-op (REQ-HEV-012, AC-HEV-015).
func TestFinalize_SelfGate_NoPendingNoOp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)

	var sink bytes.Buffer
	if err := s.FinalizeOnStop("ghost", &sink); err != nil {
		t.Fatalf("self-gate no-op should not error: %v", err)
	}
	if len(readLedger(t, dir)) != 0 {
		t.Fatal("no-op finalize must not append to the ledger")
	}
	if sink.Len() != 0 {
		t.Fatalf("no-op finalize must be silent, got: %q", sink.String())
	}
}

// TestFinalize_NonTerminalStaysPending: multi-turn pipeline survives Stop
// (REQ-HEV-012, AC-HEV-016).
func TestFinalize_NonTerminalStaysPending(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)

	if err := s.Record(recordRow("multi", "run")); err != nil {
		t.Fatal(err)
	}
	// Non-terminal evidence (gate_exit 0 without terminal flag).
	if err := s.AppendEvidence("multi", EvidenceRef{Kind: KindGateExit, Value: "0"}); err != nil {
		t.Fatal(err)
	}
	if err := s.FinalizeOnStop("multi", nil); err != nil {
		t.Fatal(err)
	}
	if len(readLedger(t, dir)) != 0 {
		t.Fatal("non-terminal evidence must not finalize the row")
	}
	if !pendingExists(dir, "multi") {
		t.Fatal("pending row must survive a non-terminal Stop")
	}

	// Now a terminal signal arrives; next Stop finalizes success.
	if err := s.AppendEvidence("multi", EvidenceRef{Kind: KindGateExit, Value: "0", Terminal: true, Ref: "go test ./..."}); err != nil {
		t.Fatal(err)
	}
	if err := s.FinalizeOnStop("multi", nil); err != nil {
		t.Fatal(err)
	}
	rows := readLedger(t, dir)
	if len(rows) != 1 || rows[0].Outcome != OutcomeSuccess {
		t.Fatalf("terminal evidence must finalize success, got %+v", rows)
	}
	if pendingExists(dir, "multi") {
		t.Fatal("finalized pending row must be deleted")
	}
}

// TestFinalize_FailOpen: an injected ledger write failure is surfaced to the
// error sink and nil is returned; the pending row survives for retry
// (REQ-HEV-015, AC-HEV-018).
func TestFinalize_FailOpen(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)

	if err := s.Record(recordRow("failopen", "run")); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendEvidence("failopen", EvidenceRef{Kind: KindAbort, Value: "killed"}); err != nil {
		t.Fatal(err)
	}
	// Inject a write failure: make the ledger PATH a directory so OpenFile fails.
	if err := os.Mkdir(s.LedgerPath(), 0o755); err != nil {
		t.Fatal(err)
	}

	var sink bytes.Buffer
	if err := s.FinalizeOnStop("failopen", &sink); err != nil {
		t.Fatalf("fail-open finalizer must return nil, got %v", err)
	}
	if sink.Len() == 0 {
		t.Fatal("fail-open finalizer must surface the error to the sink")
	}
	if !pendingExists(dir, "failopen") {
		t.Fatal("pending row must survive a failed append for later retry")
	}
}

// TestRecordDefaults verifies dispatch-time defaults (model_class, ts,
// schema_version, non-nil slices) are filled.
func TestRecordDefaults(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.Record(PendingRow{SessionID: "d1", MatchedSubcommand: "plan"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, pendingPrefix+"d1"+pendingSuffix))
	if err != nil {
		t.Fatal(err)
	}
	var p PendingRow
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	if p.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %d, want %d", p.SchemaVersion, SchemaVersion)
	}
	if p.ModelClass != "unknown" {
		t.Errorf("model_class default = %q, want unknown", p.ModelClass)
	}
	if p.TS == "" || p.CreatedAt.IsZero() {
		t.Error("ts / created_at must be filled at record time")
	}
}

// TestAppendDelegation covers the A4 delegation trajectory path (REQ-HEV-004).
func TestAppendDelegation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.Record(recordRow("dg", "run")); err != nil {
		t.Fatal(err)
	}
	blk := "scope-doc"
	if err := s.AppendDelegation("dg", Delegation{Agent: "manager-develop", CycleType: "tdd", Outcome: "fail", Blocker: &blk}); err != nil {
		t.Fatal(err)
	}
	// Finalize via terminal evidence and confirm the delegation persists into the row.
	if err := s.AppendEvidence("dg", EvidenceRef{Kind: KindAbort, Value: "blocker"}); err != nil {
		t.Fatal(err)
	}
	if err := s.FinalizeOnStop("dg", nil); err != nil {
		t.Fatal(err)
	}
	rows := readLedger(t, dir)
	if len(rows) != 1 || len(rows[0].Delegations) != 1 {
		t.Fatalf("expected 1 row with 1 delegation, got %+v", rows)
	}
	d := rows[0].Delegations[0]
	if d.Agent != "manager-develop" || d.CycleType != "tdd" || d.Blocker == nil || *d.Blocker != "scope-doc" {
		t.Fatalf("delegation not persisted correctly: %+v", d)
	}
}

// TestAppendEvidence_NoPendingNoOp: appending evidence with no pending row is a
// no-op (nothing to attach to).
func TestAppendEvidence_NoPendingNoOp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.AppendEvidence("nobody", EvidenceRef{Kind: KindGateExit, Value: "0"}); err != nil {
		t.Fatalf("evidence with no pending must be a silent no-op, got %v", err)
	}
	if pendingExists(dir, "nobody") {
		t.Fatal("no pending file should be fabricated")
	}
}

// TestNoVerbatimInStoreFiles: the sentinel raw request never lands on disk in
// the pending file or ledger — only the digest does (REQ-HEV-005, AC-HEV-009).
func TestNoVerbatimInStoreFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)

	const sentinel = "TOP-SECRET-USER-PROMPT-canary-42"
	row := PendingRow{
		SessionID:         "priv",
		MatchedSubcommand: "run",
		RequestDigest:     RequestDigest(sentinel),
		RequestClass:      ClassifyRequest(sentinel),
	}
	if err := s.Record(row); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendEvidence("priv", EvidenceRef{Kind: KindGateExit, Value: "0", Terminal: true, Ref: "go test"}); err != nil {
		t.Fatal(err)
	}
	if err := s.FinalizeOnStop("priv", nil); err != nil {
		t.Fatal(err)
	}
	// Scan every file the store produced under stateDir for the sentinel.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte(sentinel)) {
			t.Fatalf("verbatim request text leaked into %s", e.Name())
		}
	}
}
