package routing

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// ─────────────────────────────────────────────────────────────
// SPEC-HARNESS-LEARNING-EVO-001 M1 — create-if-absent + annotate
// ─────────────────────────────────────────────────────────────

// TestRecordIfAbsent_Lifecycle pins the create-once semantics (AC-HLE-001,
// REQ-HLE-003/004): the second and third calls are no-ops that must NOT clobber
// content accumulated in between, and no call ever appends a ledger row.
func TestRecordIfAbsent_Lifecycle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)

	if err := s.RecordIfAbsent(recordRow("sess-L", "plan")); err != nil {
		t.Fatal(err)
	}
	// Second call for the same session: must be a no-op.
	if err := s.RecordIfAbsent(recordRow("sess-L", "run")); err != nil {
		t.Fatal(err)
	}

	// Accumulate content, then call a third time.
	if err := s.AppendDelegation("sess-L", Delegation{Agent: "manager-develop", Outcome: "success"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendDelegation("sess-L", Delegation{Agent: "plan-auditor", Outcome: "success"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendEvidence("sess-L", EvidenceRef{Kind: KindGateExit, Value: "0"}); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordIfAbsent(recordRow("sess-L", "sync")); err != nil {
		t.Fatal(err)
	}

	// Exactly one pending file for the session, content preserved.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	pendingCount := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), pendingPrefix) && strings.HasSuffix(e.Name(), pendingSuffix) {
			pendingCount++
		}
	}
	if pendingCount != 1 {
		t.Fatalf("got %d pending files, want exactly 1", pendingCount)
	}

	p, ok, err := s.loadPending(s.pendingPath("sess-L"))
	if err != nil || !ok {
		t.Fatalf("load pending: ok=%v err=%v", ok, err)
	}
	if len(p.Delegations) != 2 {
		t.Fatalf("delegations = %d, want 2 (third RecordIfAbsent must not clobber)", len(p.Delegations))
	}
	if len(p.EvidenceRefs) != 1 {
		t.Fatalf("evidence_refs = %d, want 1 (third RecordIfAbsent must not clobber)", len(p.EvidenceRefs))
	}
	// The first writer's subcommand survives; later calls do not relabel.
	if p.MatchedSubcommand != "plan" {
		t.Fatalf("matched_subcommand = %q, want %q (create-once)", p.MatchedSubcommand, "plan")
	}
	if rows := readLedger(t, dir); len(rows) != 0 {
		t.Fatalf("RecordIfAbsent must never append a ledger row, got %d: %+v", len(rows), rows)
	}
}

// TestRecord_StillReroutesSelf pins that the pre-existing dispatch path is
// unchanged by this SPEC (AC-HLE-001, REQ-HLE-004). RecordIfAbsent must not have
// been implemented by weakening Record.
func TestRecord_StillReroutesSelf(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)

	if err := s.Record(recordRow("sess-R", "plan")); err != nil {
		t.Fatal(err)
	}
	if err := s.Record(recordRow("sess-R", "run")); err != nil {
		t.Fatal(err)
	}
	rows := readLedger(t, dir)
	if len(rows) != 1 || rows[0].Outcome != OutcomeReroute {
		t.Fatalf("Record must still reroute the prior row, got %+v", rows)
	}
}

// TestRecordIfAbsent_NoSweepOnCreatePath is the behavioral proxy for
// REQ-HLE-015 / AC-HLE-002: five stale, dead foreign rows are left completely
// untouched by RecordIfAbsent, while the same fixture under Record sweeps them.
// The contrast is what makes this a proof of absence rather than of nothing.
func TestRecordIfAbsent_NoSweepOnCreatePath(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)

	seed := func(dir string) []string {
		names := make([]string, 0, 5)
		for _, id := range []string{"f1", "f2", "f3", "f4", "f5"} {
			writeForeignPending(t, dir, id, now.Add(-48*time.Hour)) // stale by age
			names = append(names, id)
		}
		return names
	}

	t.Run("RecordIfAbsent does not sweep", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		foreign := seed(dir)
		before := make(map[string][]byte, len(foreign))
		for _, id := range foreign {
			data, err := os.ReadFile(filepath.Join(dir, pendingPrefix+id+pendingSuffix))
			if err != nil {
				t.Fatal(err)
			}
			before[id] = data
		}

		s := NewStore(dir).WithClock(fixedClock(now))
		if err := s.RecordIfAbsent(recordRow("fresh", "run")); err != nil {
			t.Fatal(err)
		}

		for _, id := range foreign {
			data, err := os.ReadFile(filepath.Join(dir, pendingPrefix+id+pendingSuffix))
			if err != nil {
				t.Fatalf("foreign pending %s must not be removed: %v", id, err)
			}
			if !bytes.Equal(data, before[id]) {
				t.Fatalf("foreign pending %s must not be modified", id)
			}
		}
		if rows := readLedger(t, dir); len(rows) != 0 {
			t.Fatalf("RecordIfAbsent must not append abort rows (sweep did not run), got %d", len(rows))
		}
	})

	t.Run("Record still sweeps the same fixture", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		seed(dir)
		s := NewStore(dir).WithClock(fixedClock(now))
		if err := s.Record(recordRow("fresh", "run")); err != nil {
			t.Fatal(err)
		}
		rows := readLedger(t, dir)
		if len(rows) != 5 {
			t.Fatalf("Record must sweep all 5 stale foreign rows, got %d", len(rows))
		}
		for _, r := range rows {
			if r.Outcome != OutcomeAbort {
				t.Fatalf("swept row outcome = %q, want abort", r.Outcome)
			}
		}
	})
}

// TestSweepStale_AgeAndLivenessGuards pins BOTH sweep guards on the dispatch
// path (AC-HLE-003, REQ-HLE-004). These are what protect a concurrent
// same-checkout session's in-flight row from a false abort.
func TestSweepStale_AgeAndLivenessGuards(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)

	// Both foreign rows are aged PAST the threshold; only liveness separates them.
	writeForeignPending(t, dir, "aged-dead", now.Add(-25*time.Hour))
	writeForeignPending(t, dir, "aged-live", now.Add(-25*time.Hour))
	if err := os.WriteFile(filepath.Join(dir, activeSessionsFile),
		[]byte(`[{"session_id":"aged-live","pid":4242}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStore(dir).WithClock(fixedClock(now))
	if err := s.Record(recordRow("third", "run")); err != nil {
		t.Fatal(err)
	}

	rows := readLedger(t, dir)
	if len(rows) != 1 {
		t.Fatalf("exactly one row must be swept, got %d: %+v", len(rows), rows)
	}
	if rows[0].SessionID != "aged-dead" || rows[0].Outcome != OutcomeAbort {
		t.Fatalf("swept row = %+v; want session=aged-dead outcome=abort", rows[0])
	}
	if pendingExists(dir, "aged-dead") {
		t.Fatal("age guard passed + not live -> pending file must be removed")
	}
	if !pendingExists(dir, "aged-live") {
		t.Fatal("liveness guard must protect the live-listed row regardless of age")
	}
}

// TestAnnotate covers the patch-only operation (AC-HLE-004, REQ-HLE-005):
// patches an existing row, leaves empty fields untouched, creates nothing, and
// finalizes nothing.
func TestAnnotate(t *testing.T) {
	t.Parallel()

	t.Run("patches set fields and leaves empty ones untouched", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		s := NewStore(dir)
		if err := s.RecordIfAbsent(PendingRow{SessionID: "an-1"}); err != nil {
			t.Fatal(err)
		}
		// Seed a mode so the later empty-mode patch has something to preserve.
		mode := "sub-agent"
		if err := s.Annotate("an-1", RoutingPatch{ModeSelected: &mode}); err != nil {
			t.Fatal(err)
		}
		sub, tier := "plan", "M"
		if err := s.Annotate("an-1", RoutingPatch{MatchedSubcommand: &sub, Tier: &tier}); err != nil {
			t.Fatal(err)
		}

		p, ok, err := s.loadPending(s.pendingPath("an-1"))
		if err != nil || !ok {
			t.Fatalf("load pending: ok=%v err=%v", ok, err)
		}
		if p.MatchedSubcommand != "plan" {
			t.Errorf("matched_subcommand = %q, want plan", p.MatchedSubcommand)
		}
		if p.Tier == nil || *p.Tier != "M" {
			t.Errorf("tier = %v, want M", p.Tier)
		}
		if p.ModeSelected == nil || *p.ModeSelected != "sub-agent" {
			t.Errorf("mode_selected = %v; an unset patch field must leave the existing value untouched", p.ModeSelected)
		}
		if rows := readLedger(t, dir); len(rows) != 0 {
			t.Fatalf("Annotate must not append a ledger row, got %d", len(rows))
		}
	})

	t.Run("no pending row is a no-op that creates nothing", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		s := NewStore(dir)
		sub := "run"
		if err := s.Annotate("ghost", RoutingPatch{MatchedSubcommand: &sub}); err != nil {
			t.Fatalf("annotate with no pending row must be a silent no-op, got %v", err)
		}
		if pendingExists(dir, "ghost") {
			t.Fatal("Annotate must not fabricate a pending row")
		}
		if rows := readLedger(t, dir); len(rows) != 0 {
			t.Fatalf("Annotate must not append a ledger row, got %d", len(rows))
		}
	})
}

// TestSchemaVersionStable pins REQ-HLE-014 / AC-HLE-005: this SPEC is additive
// within schema v1. A ledger fixture written before it still parses, and every
// row the store writes carries schema_version 1.
func TestSchemaVersionStable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// A pre-SPEC ledger fixture (schema v1, no fields from this SPEC).
	legacy := `{"schema_version":1,"ts":"2026-07-01T00:00:00Z","session_id":"legacy-1","model_class":"unknown",` +
		`"request_digest":"sha256:abcdef123456","request_class":"feature","matched_subcommand":"plan",` +
		`"mode_selected":null,"tier":null,"harness_level":null,"clarify_rounds":0,"outcome":"abort",` +
		`"loop_iterations":0,"goal_converged":null,"convergence_class":null,"delegations":[],"evidence_refs":[]}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, LedgerFileName), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStore(dir)
	if err := s.RecordIfAbsent(recordRow("new-1", "run")); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendEvidence("new-1", EvidenceRef{Kind: KindGateExit, Value: "0", Terminal: true, Ref: "go test"}); err != nil {
		t.Fatal(err)
	}
	if err := s.FinalizeOnStop("new-1", nil); err != nil {
		t.Fatal(err)
	}

	rows := readLedger(t, dir)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2 (legacy + new); a legacy row failing to parse would show here", len(rows))
	}
	for _, r := range rows {
		if r.SchemaVersion != 1 {
			t.Errorf("row %s schema_version = %d, want 1", r.SessionID, r.SchemaVersion)
		}
	}
	if rows[0].SessionID != "legacy-1" || rows[0].MatchedSubcommand != "plan" {
		t.Errorf("legacy row did not round-trip: %+v", rows[0])
	}
}
