package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// activeSessionsFile is the best-effort liveness registry consulted by the stale
// sweep (REQ-HEV-014). It lives alongside the ledger under .moai/state/.
const activeSessionsFile = "active-sessions.json"

// Store manages per-session pending routing rows and finalizes them into the
// ledger. Pending state lives at .moai/state/routing-pending-<session>.json —
// one file per session (D5 per-session isolation), so the only shared file, the
// ledger, is exclusively O_APPEND'ed (never read-modify-written). Per-session
// pending files ARE read-modify-written, which is race-free because a session
// only ever touches its own pending file.
type Store struct {
	stateDir   string
	writer     *Writer
	now        func() time.Time
	staleAfter time.Duration
}

// NewStore constructs a Store rooted at stateDir (typically <root>/.moai/state).
func NewStore(stateDir string) *Store {
	return &Store{
		stateDir:   stateDir,
		writer:     NewWriter(filepath.Join(stateDir, LedgerFileName)),
		now:        func() time.Time { return time.Now().UTC() },
		staleAfter: StalenessThreshold,
	}
}

// WithClock overrides the Store clock (test seam). Returns the Store for
// chaining. The clock supplies both the created-at stamp and the sweep age
// reference, so age-guarded behavior is deterministically testable without
// wall-clock sleeps (AP-5 — no timing assertions).
func (s *Store) WithClock(now func() time.Time) *Store {
	s.now = now
	return s
}

// LedgerPath returns the absolute ledger path this Store appends to.
func (s *Store) LedgerPath() string { return filepath.Join(s.stateDir, LedgerFileName) }

// sessionKey maps an empty/unresolvable session id to the degraded single-session
// key so pending state is still functional (D5).
func sessionKey(sessionID string) string {
	if sessionID == "" {
		return "nosession"
	}
	return sessionID
}

func (s *Store) pendingPath(sessionID string) string {
	return filepath.Join(s.stateDir, pendingPrefix+sessionKey(sessionID)+pendingSuffix)
}

// Record creates a new pending row for row.SessionID after (1) sweeping stale
// FOREIGN rows as abort and (2) rerouting this session's own prior pending row
// (REQ-HEV-011). Dispatch-time defaults are filled here.
func (s *Store) Record(row PendingRow) error {
	row.SchemaVersion = SchemaVersion
	if row.CreatedAt.IsZero() {
		row.CreatedAt = s.now().UTC()
	}
	if row.TS == "" {
		row.TS = row.CreatedAt.UTC().Format(time.RFC3339)
	}
	if row.ModelClass == "" {
		row.ModelClass = "unknown"
	}
	if row.Delegations == nil {
		row.Delegations = []Delegation{}
	}
	if row.EvidenceRefs == nil {
		row.EvidenceRefs = []EvidenceRef{}
	}

	// (1) Age-guarded, liveness-guarded stale sweep of FOREIGN rows.
	if err := s.sweepStale(row.SessionID); err != nil {
		return err
	}
	// (2) Same-session reroute (age-independent — REQ-HEV-010 precedence: the
	//     current session's own row is ALWAYS rerouted, never swept).
	if err := s.rerouteSelf(row.SessionID); err != nil {
		return err
	}
	// (3) Persist the fresh pending row.
	return s.writePending(row)
}

// AppendEvidence appends a machine evidence ref to the session's pending row.
// No pending row for the session -> no-op (evidence with nothing to attach to is
// dropped rather than fabricating a row).
func (s *Store) AppendEvidence(sessionID string, ref EvidenceRef) error {
	pending, ok, err := s.loadPending(s.pendingPath(sessionID))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	pending.EvidenceRefs = append(pending.EvidenceRefs, ref)
	return s.writePending(pending)
}

// AppendDelegation appends a subagent delegation trajectory entry (design delta
// A4) to the session's pending row. No pending row -> no-op.
func (s *Store) AppendDelegation(sessionID string, d Delegation) error {
	pending, ok, err := s.loadPending(s.pendingPath(sessionID))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	pending.Delegations = append(pending.Delegations, d)
	return s.writePending(pending)
}

// FinalizeOnStop is the Stop-hook self-gated finalizer (REQ-HEV-012). It is
// fail-open (REQ-HEV-015): every error is written to errSink (when non-nil) and
// nil is returned — the finalizer NEVER blocks session end.
//
//   - no pending row for the session  -> no-op (self-gate; the Stop hook fires
//     every turn-end, so a routing-less turn must be a clean no-op)
//   - DeriveOutcome non-terminal        -> leave pending (multi-turn pipeline)
//   - terminal                          -> append the finalized row + delete the
//     pending file
//
// A failed ledger append leaves the pending file in place so a later Stop can
// retry, rather than silently losing the observation.
func (s *Store) FinalizeOnStop(sessionID string, errSink io.Writer) error {
	path := s.pendingPath(sessionID)
	pending, ok, err := s.loadPending(path)
	if err != nil {
		logErr(errSink, "routing finalize: load pending: %v", err)
		return nil
	}
	if !ok {
		return nil // self-gate: no pending row -> no-op
	}
	outcome, terminal := DeriveOutcome(pending.EvidenceRefs)
	if !terminal {
		return nil // multi-turn pipeline: stay pending
	}
	if err := s.writer.Append(pending.Finalize(outcome)); err != nil {
		logErr(errSink, "routing finalize: append ledger: %v", err)
		return nil // fail-open: keep pending for a later retry
	}
	if err := os.Remove(path); err != nil {
		logErr(errSink, "routing finalize: remove pending: %v", err)
	}
	return nil
}

// rerouteSelf finalizes this session's own prior pending row as reroute
// (REQ-HEV-010), bypassing DeriveOutcome. A malformed own-pending file is
// dropped so the fresh row can overwrite it.
func (s *Store) rerouteSelf(sessionID string) error {
	path := s.pendingPath(sessionID)
	pending, ok, err := s.loadPending(path)
	if err != nil {
		_ = os.Remove(path) // unparseable own row -> drop and let writePending overwrite
		return nil
	}
	if !ok {
		return nil
	}
	if err := s.writer.Append(pending.Finalize(OutcomeReroute)); err != nil {
		return err
	}
	return os.Remove(path)
}

// sweepStale finalizes FOREIGN or unresolvable-session pending rows as abort
// (REQ-HEV-014), bypassing DeriveOutcome, subject to BOTH guards:
//
//   - age guard: only rows OLDER than staleAfter are swept; a foreign row at or
//     younger than the threshold is left untouched (a live parallel same-checkout
//     session's in-flight row is documented-normal in this repo and must not be
//     falsely aborted).
//   - liveness guard: a row whose session_id is listed live in active-sessions.json
//     is never swept regardless of age (best-effort; file absent/unreadable ->
//     fall back to the age rule alone).
//
// The current session's own row is never swept here — it is excluded by key and
// handled by rerouteSelf.
func (s *Store) sweepStale(currentSession string) error {
	entries, err := os.ReadDir(s.stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	currentKey := sessionKey(currentSession)
	now := s.now().UTC()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, pendingPrefix) || !strings.HasSuffix(name, pendingSuffix) {
			continue
		}
		key := strings.TrimSuffix(strings.TrimPrefix(name, pendingPrefix), pendingSuffix)
		if key == currentKey {
			continue // same session -> rerouteSelf handles it, never swept
		}
		path := filepath.Join(s.stateDir, name)
		pending, ok, loadErr := s.loadPending(path)
		if loadErr != nil || !ok {
			continue // unreadable/malformed foreign file -> leave it (fail-open)
		}
		if s.isSessionLive(pending.SessionID) {
			continue // liveness guard: never abort a live-listed session's row
		}
		if now.Sub(pending.CreatedAt.UTC()) <= s.staleAfter {
			continue // age guard: young foreign row -> untouched
		}
		if err := s.writer.Append(pending.Finalize(OutcomeAbort)); err != nil {
			return err
		}
		_ = os.Remove(path)
	}
	return nil
}

// isSessionLive reports whether sessionID appears in active-sessions.json.
// Best-effort: an unresolvable id, or an absent/unreadable registry, returns
// false (the sweep then falls back to the age rule alone). Substring containment
// is sufficient because session ids are long unique UUIDs, and erring toward
// "not live" only ever DEFERS a sweep — it never falsely aborts a live row.
func (s *Store) isSessionLive(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(s.stateDir, activeSessionsFile))
	if err != nil {
		return false
	}
	return bytes.Contains(data, []byte(sessionID))
}

// loadPending reads and parses a pending-row file. The bool is false (with a nil
// error) when the file does not exist.
func (s *Store) loadPending(path string) (PendingRow, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PendingRow{}, false, nil
		}
		return PendingRow{}, false, err
	}
	var p PendingRow
	if err := json.Unmarshal(data, &p); err != nil {
		return PendingRow{}, false, fmt.Errorf("routing: parse pending %s: %w", filepath.Base(path), err)
	}
	return p, true, nil
}

// writePending persists a pending row to its per-session file.
func (s *Store) writePending(row PendingRow) error {
	if err := os.MkdirAll(s.stateDir, 0o755); err != nil {
		return fmt.Errorf("routing: ensure state dir: %w", err)
	}
	data, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("routing: marshal pending: %w", err)
	}
	if err := os.WriteFile(s.pendingPath(row.SessionID), data, 0o644); err != nil {
		return fmt.Errorf("routing: write pending: %w", err)
	}
	return nil
}

// logErr writes a fail-open diagnostic to sink when sink is non-nil.
func logErr(sink io.Writer, format string, args ...any) {
	if sink == nil {
		return
	}
	_, _ = fmt.Fprintf(sink, format+"\n", args...)
}
