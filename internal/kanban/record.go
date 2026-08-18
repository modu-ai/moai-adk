// Package kanban implements the state record that carries Kanban Mode's
// session-scoped signal between the launcher that opens a kanban session and
// the orchestrator that drives the chain inside it.
//
// The record is best-effort by design: a launch never depends on it. A session
// whose record could not be written is simply a session with no record, which
// downstream reads as "no evidence" and resolves in the safe direction — run
// the check rather than skip it.
package kanban

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Session backends a kanban chain can run on. A mixed-backend session is not a
// kanban session, so no third value exists.
const (
	BackendClaude = "claude"
	BackendGLM    = "glm"
)

// Rung is the rigor rung a verify result self-labels with. It is orthogonal to
// the result's severity: a rung describes how rigorously the scan ran, never
// whether it found anything.
type Rung string

// The three rungs a readable verify result can carry. A verify stage that
// produced no readable result carries no rung at all — that is the absent case,
// not a fourth rung.
const (
	RungPrimary  Rung = "PRIMARY"
	RungFallback Rung = "FALLBACK"
	RungDegraded Rung = "DEGRADED"
)

// stateDirSegments is the record's home beneath the project root, matching the
// per-subsystem layout the rest of .moai/state/ already uses.
var stateDirSegments = []string{".moai", "state", "kanban"}

// @MX:ANCHOR: [AUTO] Kanban state record schema — the cross-actor contract for a kanban session
// @MX:REASON: the launcher writes SessionID/SpecID/Backend/EnteredAt at launch while the orchestrator fills DeepScanDir/VerifyRung/VerifyReentries later; both sides plus the sync-phase dedup gate bind to these JSON keys, so a renamed key breaks readers this package cannot see
//
// Record is the per-session kanban state record persisted at
// .moai/state/kanban/<session>.json.
//
// The three orchestrator-written fields (DeepScanDir, VerifyRung,
// VerifyReentries) are filled in independently as the chain progresses, so a
// record carrying some but not others is a normal intermediate state rather
// than corruption.
type Record struct {
	// SessionID keys the record and names its file. Required.
	SessionID string `json:"session_id"`

	// SpecID names the SPEC the chain targets. Empty is legitimate: it means
	// the chain heads at plan-phase from the operator's first prompt.
	SpecID string `json:"spec_id"`

	// Role is the chain role this session occupies: leader | planner | runner
	// | syncer, or "worker" for a factory run's numbered lane. It is derived
	// from the companion label (the bare role name, or its bumped
	// `<role>-<n>` form) or the factory worker label (`worker-<n>`) at
	// launch, or "leader" for the session that elected the run.
	//
	// Empty is legitimate and load-bearing: a record written before this field
	// existed, or a launch whose label could not be parsed, leaves it blank —
	// and a consumer MUST render that as "not recorded" rather than guessing a
	// role. omitempty keeps pre-existing records byte-identical on rewrite.
	Role string `json:"role,omitempty"`

	// Backend is BackendClaude or BackendGLM.
	Backend string `json:"backend"`

	// EnteredAt is the RFC3339 instant the session entered Kanban Mode.
	EnteredAt string `json:"entered_at"`

	// DeepScanDir is the verify stage's results directory, written by the
	// orchestrator once a scan has produced one.
	DeepScanDir string `json:"deepscan_dir"`

	// VerifyRung is the rung of the most recent readable verify result.
	//
	// It is a pointer so that "never recorded" (nil) stays distinguishable from
	// "recorded as empty" (a pointer to ""). That distinction is load-bearing:
	// because this record is best-effort and its three orchestrator-written
	// fields land independently, a record carrying DeepScanDir but no rung is
	// reachable. Collapsing absent into empty would leave a consumer unable to
	// tell a rung it never received from one it received blank — and the
	// downstream suppression check is an allow-list over a *recorded* PRIMARY
	// or FALLBACK, so both cases must read as "no suppression" on evidence
	// rather than by accident.
	VerifyRung *Rung `json:"verify_rung,omitempty"`

	// VerifyReentries counts verify re-entries consumed so far.
	VerifyReentries int `json:"verify_reentries"`
}

// NewRecord builds a record for a session entering Kanban Mode, stamping
// EnteredAt with the current UTC instant in RFC3339. The three
// orchestrator-written fields are deliberately left at their zero values —
// VerifyRung nil rather than empty — so a reader can tell they have not been
// written yet.
func NewRecord(sessionID, specID, backend string) *Record {
	return &Record{
		SessionID: sessionID,
		SpecID:    specID,
		Backend:   backend,
		EnteredAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// WithRole returns rec with the chain role attached. A role outside the known
// set is discarded rather than stored, so a consumer never has to defend
// against an arbitrary string arriving from a launch label. The known set is
// the kanban roles (leader + the three companions) plus RoleWorker, which a
// factory run's numbered workers record (SPEC-FACTORY-WORKER-FANOUT-001).
func (r *Record) WithRole(role string) *Record {
	if r == nil {
		return nil
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role == RoleLead || role == RoleWorker || isCompanionRole(role) {
		r.Role = role
	}
	return r
}

// RecordPath returns the on-disk path of a session's record.
func RecordPath(projectRoot, sessionID string) string {
	segments := append([]string{projectRoot}, stateDirSegments...)
	return filepath.Join(append(segments, sessionID+".json")...)
}

// Write persists rec under projectRoot, creating the state directory as needed.
//
// It reports failures so tests and tooling can observe them, but a launch path
// MUST NOT gate on the returned error — use WriteBestEffort there, whose
// signature makes gating impossible.
func Write(projectRoot string, rec *Record) error {
	if rec == nil {
		return fmt.Errorf("write kanban record: record is nil")
	}
	if err := validateSessionID(rec.SessionID); err != nil {
		return fmt.Errorf("write kanban record: %w", err)
	}

	path := RecordPath(projectRoot, rec.SessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("write kanban record: creating state directory: %w", err)
	}

	encoded, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("write kanban record: encoding: %w", err)
	}
	encoded = append(encoded, '\n')

	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return fmt.Errorf("write kanban record: %w", err)
	}
	return nil
}

// @MX:NOTE: [AUTO] the absent error return is the fail-open guarantee, not an oversight — a launch path that cannot observe a failure cannot block on one
//
// WriteBestEffort persists rec and discards every failure. It is the call the
// launcher makes: a record is an aid to the chain, never a precondition for
// starting one, and a session that launches without a record degrades to a
// non-kanban session rather than failing to launch.
func WriteBestEffort(projectRoot string, rec *Record) {
	_ = Write(projectRoot, rec)
}

// Read loads a session's record. An absent, unreadable, or malformed record is
// an error, never a zero-valued record — a caller must be able to tell "no
// record" from "a record that says nothing".
func Read(projectRoot, sessionID string) (*Record, error) {
	if err := validateSessionID(sessionID); err != nil {
		return nil, fmt.Errorf("read kanban record: %w", err)
	}

	raw, err := os.ReadFile(RecordPath(projectRoot, sessionID))
	if err != nil {
		return nil, fmt.Errorf("read kanban record: %w", err)
	}

	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, fmt.Errorf("read kanban record: decoding: %w", err)
	}
	return &rec, nil
}

// validateSessionID rejects identifiers that cannot safely name a file. The
// record path is derived from the session id, so a separator or a parent
// reference would place the record outside the state directory.
func validateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session id is empty")
	}
	if sessionID == "." || sessionID == ".." {
		return fmt.Errorf("session id %q is a directory reference", sessionID)
	}
	if strings.ContainsAny(sessionID, `/\`) {
		return fmt.Errorf("session id %q contains a path separator", sessionID)
	}
	return nil
}
