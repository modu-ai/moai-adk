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

// The record's home beneath the project root is the shared state directory
// resolved by state_dir.go — the SAME directory the queue lives in, which is
// why the two travel together through the rename (REQ-TOSQ-015). RecordPath
// resolves it purely: reading or writing a session record must not trigger the
// one-time relocation, which belongs to the `moai todo` command path.

// @MX:ANCHOR: [AUTO] Kanban state record schema — the cross-actor contract for a kanban session
// @MX:REASON: the launcher writes SessionID/SpecID/Backend/EnteredAt at launch while the orchestrator fills DeepScanDir/VerifyRung/VerifyReentries later; both sides plus the sync-phase dedup gate bind to these JSON keys, so a renamed key breaks readers this package cannot see
//
// Record is the per-session kanban state record persisted at
// <state-dir>/<session>.json (see state_dir.go for the directory).
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

	// Role is the chain role this session occupies: lead | plan | run
	// | sync, or "lane" for a factory run's numbered lane. It is derived
	// from the companion label (the bare role name, or its bumped
	// `<role>-<n>` form) or the factory lane label (`lane-<n>`) at
	// launch, or "lead" for the session that elected the run.
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

	// Lane is a factory lane's number, and 0 means "not a lane". Lanes number
	// from 1, so the zero value is unreachable by legitimate data and carries
	// the distinction a pointer would otherwise be needed for.
	//
	// It is data BESIDE Role, never a role value: Role stays "lane" for every
	// lane, and the number is what makes N lanes distinguishable. It
	// deliberately does not travel through WithRole, whose drop-unknown guard
	// exists so a consumer never has to defend against arbitrary launch-label
	// text (SPEC-KANBAN-RECORD-SESSION-KEY-001 REQ-KRS-004).
	Lane int `json:"lane,omitempty"`

	// CardID is the queue card the session is working, distinct from SpecID
	// because Class A and Class B cards never acquire a SPEC identifier.
	//
	// Empty means "not recorded" and is the correct answer rather than a
	// degraded one: the value is taken from an explicit launch override, and
	// otherwise from the session's worktree-root basename ONLY where that
	// root's parent directory is named `worktrees`. Outside a card worktree
	// the basename is a checkout name, not a card, so the field is left empty
	// rather than guessed (REQ-KRS-005).
	CardID string `json:"card_id,omitempty"`
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
// the kanban roles (lead + the three companions) plus RoleLane, which a
// factory run's numbered lanes record (SPEC-FACTORY-WORKER-FANOUT-001).
func (r *Record) WithRole(role string) *Record {
	if r == nil {
		return nil
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role == RoleLead || role == RoleLane || isCompanionRole(role) {
		r.Role = role
	}
	return r
}

// WithLane returns rec carrying a factory lane's number. A non-positive
// number is ignored rather than stored: lanes number from 1, so 0 stays the
// "not a lane" signal instead of being overwritten by an unparsed label.
//
// This is deliberately NOT part of WithRole. Widening the role setter to
// pattern-match a `lane-<n>` label would reopen exactly what its drop-unknown
// guard closes — arbitrary launch-label text reaching the role field.
func (r *Record) WithLane(lane int) *Record {
	if r == nil {
		return nil
	}
	if lane > 0 {
		r.Lane = lane
	}
	return r
}

// WithCard returns rec carrying the queue card identifier the session is
// working. An empty value leaves the field unset — "not recorded" is the
// correct reading, and a consumer must render it as that rather than guessing
// a card.
func (r *Record) WithCard(cardID string) *Record {
	if r == nil {
		return nil
	}
	if cardID = strings.TrimSpace(cardID); cardID != "" {
		r.CardID = cardID
	}
	return r
}

// RecordPath returns the on-disk path of a session's record.
func RecordPath(projectRoot, sessionID string) string {
	dir, _ := resolveStateDir(projectRoot, false)
	return filepath.Join(dir, sessionID+".json")
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
