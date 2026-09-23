package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"time"
)

// Lane handoff states (SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 design §3).
const (
	HandoffReserved                 = "RESERVED"
	HandoffWTReady                  = "WT_READY"
	HandoffSwitchPendingInteractive = "SWITCH_PENDING_INTERACTIVE"
	HandoffSwitchPendingHeadless    = "SWITCH_PENDING_HEADLESS"
	HandoffBound                    = "BOUND"
	HandoffNack                     = "NACK"
	HandoffAbandoned                = "ABANDONED"

	HandoffModeInteractive = "interactive"
	HandoffModeHeadless    = "headless"
)

// Handoff NACK reasons. Admission reasons are decided either inside the
// broker write transaction (endpoint, in-flight, stale) or by the controller
// from filesystem and Git facts before any reservation is written.
const (
	NackEndpointLaunchPending = "ENDPOINT_LAUNCH_PENDING"
	NackLaneUnknown           = "LANE_UNKNOWN"
	NackHandoffInFlight       = "HANDOFF_IN_FLIGHT"
	NackStaleReservation      = "STALE_RESERVATION"
	NackInvalidRequest        = "INVALID_REQUEST"
	NackLaneActiveTurn        = "LANE_ACTIVE_TURN"
	NackLanePermissionWait    = "LANE_PERMISSION_WAIT"
	NackLaneInterrupting      = "LANE_INTERRUPTING"
	NackSourceDirty           = "SOURCE_DIRTY"
	NackUntrustedCwd          = "UNTRUSTED_CWD"
	NackBranchCollision       = "BRANCH_COLLISION"
	NackTargetPathConflict    = "TARGET_PATH_CONFLICT"
	NackTargetDirty           = "TARGET_DIRTY"
	NackBaseDrift             = "BASE_DRIFT"

	// Relocation reasons (M2). A readback mismatch is found before any
	// app-server request; the other two after a headless request was issued.
	NackTargetReadbackMismatch    = "TARGET_READBACK_MISMATCH"
	NackRelocationRPCFailed       = "RELOCATION_RPC_FAILED"
	NackRelocationEvidenceInvalid = "RELOCATION_EVIDENCE_INVALID"
)

// Official Codex app-server methods whose result is headless relocation
// evidence (REQ-FLH-007). Nothing else is: not SessionStart, not an empty
// turn/start, not turn/steer.
const (
	RelocationMethodThreadFork  = "thread/fork"
	RelocationMethodThreadStart = "thread/start"
)

// handoffSchema extends the existing broker; it never creates a second store.
// The partial unique index makes "one unfinished handoff per lane" a storage
// invariant, not only a check.
const handoffSchema = `
CREATE TABLE IF NOT EXISTS lane_handoffs(id TEXT PRIMARY KEY, project_key TEXT NOT NULL, run_id TEXT NOT NULL, slot TEXT NOT NULL, card_id TEXT NOT NULL, spec_id TEXT NOT NULL, mode TEXT NOT NULL, nonce TEXT NOT NULL UNIQUE, handoff_generation INTEGER NOT NULL, source_backend TEXT NOT NULL, source_role TEXT NOT NULL, source_session TEXT NOT NULL, source_generation INTEGER NOT NULL, source_pid INTEGER NOT NULL, source_process_start TEXT NOT NULL, develop_pin TEXT NOT NULL, target_path TEXT NOT NULL, target_branch TEXT NOT NULL, state TEXT NOT NULL, reason TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE UNIQUE INDEX IF NOT EXISTS lane_handoffs_one_open ON lane_handoffs(slot) WHERE state IN ('RESERVED','WT_READY','SWITCH_PENDING_INTERACTIVE','SWITCH_PENDING_HEADLESS');
CREATE TABLE IF NOT EXISTS lane_handoff_events(id INTEGER PRIMARY KEY AUTOINCREMENT, handoff_id TEXT NOT NULL, from_state TEXT NOT NULL, to_state TEXT NOT NULL, reason TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS lane_handoff_relocations(handoff_id TEXT PRIMARY KEY, nonce TEXT NOT NULL, method TEXT NOT NULL, source_thread_id TEXT NOT NULL, thread_id TEXT NOT NULL, forked_from_id TEXT NOT NULL, thread_started INTEGER NOT NULL, request_cwd TEXT NOT NULL, response_cwd TEXT NOT NULL, readback_cwd TEXT NOT NULL, readback_branch TEXT NOT NULL, readback_head TEXT NOT NULL, recorded_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS lane_endpoint_tombstones(slot TEXT NOT NULL, session_uuid TEXT NOT NULL, generation INTEGER NOT NULL, replaced_by_session TEXT NOT NULL, replaced_by_generation INTEGER NOT NULL, handoff_id TEXT NOT NULL, bound_at TEXT NOT NULL, PRIMARY KEY(slot,session_uuid,generation));
`

var (
	handoffCardID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)
	handoffBranch = regexp.MustCompile(`^WT-[a-z0-9]+(-[a-z0-9]+)*$`)
	handoffPin    = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)
)

// HandoffNackError is a fail-closed handoff refusal with a stable reason code.
type HandoffNackError struct{ Reason, Detail string }

func (e *HandoffNackError) Error() string {
	if e.Detail == "" {
		return "handoff NACK " + e.Reason
	}
	return "handoff NACK " + e.Reason + ": " + e.Detail
}

// NewHandoffNack builds a NACK error for callers outside the broker.
func NewHandoffNack(reason, detail string) error {
	return &HandoffNackError{Reason: reason, Detail: detail}
}

// HandoffNackReason reports the NACK reason carried by err, if any.
func HandoffNackReason(err error) (string, bool) {
	var nack *HandoffNackError
	if errors.As(err, &nack) {
		return nack.Reason, true
	}
	if stale, ok := StaleEndpoint(err); ok {
		return stale.Code, true
	}
	return "", false
}

// HandoffReservation is a reservation request. ExpectedSource, when set, is
// the endpoint tuple the caller resolved; it must still be current inside the
// reservation transaction.
type HandoffReservation struct {
	Slot, CardID, SpecID, Mode           string
	DevelopPin, TargetPath, TargetBranch string
	ExpectedSource                       *Peer
}

// Handoff is one stored lane handoff; Source is the reserved endpoint tuple.
type Handoff struct {
	ID, ProjectKey, RunID, Slot, CardID, SpecID, Mode, Nonce string
	Generation                                               int64
	Source                                                   Peer
	DevelopPin, TargetPath, TargetBranch                     string
	State, Reason                                            string
	CreatedAt, UpdatedAt                                     time.Time
}

func (r HandoffReservation) validate() error {
	switch {
	case !safeID.MatchString(r.Slot):
		return NewHandoffNack(NackInvalidRequest, "invalid lane")
	case !handoffCardID.MatchString(r.CardID):
		return NewHandoffNack(NackInvalidRequest, "invalid card id")
	case !safeID.MatchString(r.SpecID):
		return NewHandoffNack(NackInvalidRequest, "invalid SPEC id")
	case r.Mode != HandoffModeInteractive && r.Mode != HandoffModeHeadless:
		return NewHandoffNack(NackInvalidRequest, "invalid mode")
	case !handoffPin.MatchString(r.DevelopPin):
		return NewHandoffNack(NackInvalidRequest, "invalid develop pin")
	case !filepath.IsAbs(r.TargetPath) || filepath.Clean(r.TargetPath) != r.TargetPath:
		return NewHandoffNack(NackInvalidRequest, "target path must be absolute and clean")
	case !handoffBranch.MatchString(r.TargetBranch) || len(r.TargetBranch) > 27:
		return NewHandoffNack(NackInvalidRequest, "invalid target branch")
	}
	return nil
}

func isOpenHandoffState(state string) bool {
	switch state {
	case HandoffReserved, HandoffWTReady, HandoffSwitchPendingInteractive, HandoffSwitchPendingHeadless:
		return true
	}
	return false
}

// ReserveHandoff admits a handoff for a lane and pins its source endpoint.
// Admission — reading the lane's endpoint row, the launch-pending check, the
// stale and in-flight checks, and the RESERVED insert — runs inside ONE broker
// write transaction (BEGIN IMMEDIATE, from the _txlock=immediate DSN), so a
// concurrent launcher registration is either fully before or fully after it.
//
// @MX:WARN: [AUTO] admission reads the endpoint row inside its own write transaction
// @MX:REASON: REQ-FLH-016 / AC-FLH-019 (vii) — reading the row before BEGIN pins a stale tuple once a concurrent launcher registration commits
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func (s *Store) ReserveHandoff(ctx context.Context, r HandoffReservation) (Handoff, error) {
	if err := r.validate(); err != nil {
		return Handoff{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Handoff{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var src Peer
	err = tx.QueryRowContext(ctx, `SELECT project_key,run_id,backend,role,slot,session_uuid,generation,pid,process_start FROM peers WHERE slot=?`, r.Slot).Scan(
		&src.ProjectKey, &src.RunID, &src.Backend, &src.Role, &src.Slot, &src.SessionUUID, &src.Generation, &src.PID, &src.ProcessStart)
	if errors.Is(err, sql.ErrNoRows) {
		return Handoff{}, NewHandoffNack(NackLaneUnknown, r.Slot)
	}
	if err != nil {
		return Handoff{}, err
	}
	if isLaunchPendingSession(src.SessionUUID) {
		return Handoff{}, NewHandoffNack(NackEndpointLaunchPending, r.Slot)
	}
	if e := r.ExpectedSource; e != nil && (e.SessionUUID != src.SessionUUID || e.Generation != src.Generation || e.PID != src.PID || e.ProcessStart != src.ProcessStart) {
		return Handoff{}, NewHandoffNack(NackStaleReservation, fmt.Sprintf("lane %s is at generation %d", r.Slot, src.Generation))
	}
	var open int
	var lastGen int64
	if err := tx.QueryRowContext(ctx, `SELECT coalesce(sum(state IN ('RESERVED','WT_READY','SWITCH_PENDING_INTERACTIVE','SWITCH_PENDING_HEADLESS')),0), coalesce(max(handoff_generation),0) FROM lane_handoffs WHERE slot=?`, r.Slot).Scan(&open, &lastGen); err != nil {
		return Handoff{}, err
	}
	if open > 0 {
		return Handoff{}, NewHandoffNack(NackHandoffInFlight, r.Slot)
	}

	now := s.now().UTC()
	h := Handoff{
		ID: newID(), ProjectKey: s.projectKey, RunID: s.runID, Slot: r.Slot, CardID: r.CardID, SpecID: r.SpecID,
		Mode: r.Mode, Nonce: newID(), Generation: lastGen + 1, Source: src,
		DevelopPin: r.DevelopPin, TargetPath: r.TargetPath, TargetBranch: r.TargetBranch,
		State: HandoffReserved, CreatedAt: now, UpdatedAt: now,
	}
	stamp := now.Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO lane_handoffs(id,project_key,run_id,slot,card_id,spec_id,mode,nonce,handoff_generation,source_backend,source_role,source_session,source_generation,source_pid,source_process_start,develop_pin,target_path,target_branch,state,reason,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'',?,?)`,
		h.ID, h.ProjectKey, h.RunID, h.Slot, h.CardID, h.SpecID, h.Mode, h.Nonce, h.Generation,
		src.Backend, src.Role, src.SessionUUID, src.Generation, src.PID, src.ProcessStart,
		h.DevelopPin, h.TargetPath, h.TargetBranch, h.State, stamp, stamp); err != nil {
		return Handoff{}, err
	}
	if err := insertHandoffEvent(ctx, tx, h.ID, "", HandoffReserved, "", stamp); err != nil {
		return Handoff{}, err
	}
	if err := tx.Commit(); err != nil {
		return Handoff{}, err
	}
	return h, nil
}

func insertHandoffEvent(ctx context.Context, tx *sql.Tx, id, from, to, reason, at string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO lane_handoff_events(handoff_id,from_state,to_state,reason,created_at) VALUES(?,?,?,?,?)`, id, from, to, reason, at)
	return err
}

// MarkHandoffWTReady records a verified target worktree for a RESERVED handoff.
func (s *Store) MarkHandoffWTReady(ctx context.Context, h Handoff) (Handoff, error) {
	return s.transitionHandoff(ctx, h, func(state, _ string) bool { return state == HandoffReserved }, HandoffWTReady, "")
}

// MarkHandoffSwitchPendingInteractive records that the user-executed /cd
// guidance was issued for an interactive WT_READY handoff (REQ-FLH-006).
func (s *Store) MarkHandoffSwitchPendingInteractive(ctx context.Context, h Handoff) (Handoff, error) {
	return s.transitionHandoff(ctx, h, switchFrom(HandoffModeInteractive), HandoffSwitchPendingInteractive, "")
}

// MarkHandoffSwitchPendingHeadless records that the official app-server
// relocation request is about to be issued for a headless WT_READY handoff
// (REQ-FLH-007). The two modes never enter each other's pending state.
func (s *Store) MarkHandoffSwitchPendingHeadless(ctx context.Context, h Handoff) (Handoff, error) {
	return s.transitionHandoff(ctx, h, switchFrom(HandoffModeHeadless), HandoffSwitchPendingHeadless, "")
}

func switchFrom(mode string) func(state, storedMode string) bool {
	return func(state, storedMode string) bool { return state == HandoffWTReady && storedMode == mode }
}

// NackHandoff terminally refuses an unfinished handoff; only a fresh
// reservation can follow it.
func (s *Store) NackHandoff(ctx context.Context, h Handoff, reason string) (Handoff, error) {
	if reason == "" {
		return Handoff{}, errors.New("handoff NACK requires a reason")
	}
	return s.transitionHandoff(ctx, h, func(state, _ string) bool { return isOpenHandoffState(state) }, HandoffNack, reason)
}

// transitionHandoff is a nonce-and-state CAS; a stale caller changes nothing.
func (s *Store) transitionHandoff(ctx context.Context, h Handoff, allowedFrom func(state, mode string) bool, to, reason string) (Handoff, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Handoff{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var state, mode string
	if err := tx.QueryRowContext(ctx, `SELECT state,mode FROM lane_handoffs WHERE id=? AND nonce=?`, h.ID, h.Nonce).Scan(&state, &mode); err != nil {
		return Handoff{}, fmt.Errorf("handoff %s: %w", h.ID, err)
	}
	if !allowedFrom(state, mode) {
		return Handoff{}, fmt.Errorf("handoff %s cannot move %s -> %s", h.ID, state, to)
	}
	now := s.now().UTC()
	stamp := now.Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `UPDATE lane_handoffs SET state=?,reason=?,updated_at=? WHERE id=? AND nonce=? AND state=?`, to, reason, stamp, h.ID, h.Nonce, state); err != nil {
		return Handoff{}, err
	}
	if err := insertHandoffEvent(ctx, tx, h.ID, state, to, reason, stamp); err != nil {
		return Handoff{}, err
	}
	if err := tx.Commit(); err != nil {
		return Handoff{}, err
	}
	h.State, h.Reason, h.UpdatedAt = to, reason, now
	return h, nil
}

// HandoffsForLane lists a lane's handoffs, oldest first.
func (s *Store) HandoffsForLane(ctx context.Context, slot string) (_ []Handoff, err error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_key,run_id,slot,card_id,spec_id,mode,nonce,handoff_generation,source_backend,source_role,source_session,source_generation,source_pid,source_process_start,develop_pin,target_path,target_branch,state,reason,created_at,updated_at FROM lane_handoffs WHERE slot=? ORDER BY handoff_generation`, slot)
	if err != nil {
		return nil, err
	}
	defer closeInto(&err, rows, "handoff rows")
	var out []Handoff
	for rows.Next() {
		var h Handoff
		var created, updated string
		if err := rows.Scan(&h.ID, &h.ProjectKey, &h.RunID, &h.Slot, &h.CardID, &h.SpecID, &h.Mode, &h.Nonce, &h.Generation,
			&h.Source.Backend, &h.Source.Role, &h.Source.SessionUUID, &h.Source.Generation, &h.Source.PID, &h.Source.ProcessStart,
			&h.DevelopPin, &h.TargetPath, &h.TargetBranch, &h.State, &h.Reason, &created, &updated); err != nil {
			return nil, err
		}
		h.Source.ProjectKey, h.Source.RunID, h.Source.Slot = h.ProjectKey, h.RunID, h.Slot
		h.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		h.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		out = append(out, h)
	}
	return out, rows.Err()
}
