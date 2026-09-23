package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// HeadlessRelocation is the evidence a headless handoff carries into the M3
// atomic rebind (REQ-FLH-007/008): the official app-server result (method,
// returned thread id, lineage, observed thread/started, reported cwd) and the
// controller's own readback of the target (cwd, branch, HEAD). Recording it
// changes no endpoint: BOUND is M3's single transaction.
type HeadlessRelocation struct {
	HandoffID, Nonce                          string
	Method                                    string
	SourceThreadID, ThreadID, ForkedFromID    string
	ThreadStarted                             bool
	RequestCwd, ResponseCwd                   string
	ReadbackCwd, ReadbackBranch, ReadbackHead string
	RecordedAt                                time.Time
}

func relocationNack(detail string) error {
	return &HandoffNackError{Reason: NackRelocationEvidenceInvalid, Detail: detail}
}

// validateRelocation accepts only official relocation evidence for h. A stored
// source thread demands a history-preserving thread/fork whose lineage names
// it; no stored history demands thread/start. SessionStart, an empty
// turn/start, and turn/steer are named explicitly so a caller that tries to
// manufacture binding evidence gets a precise refusal.
func validateRelocation(h Handoff, r HeadlessRelocation) error {
	switch r.Method {
	case RelocationMethodThreadFork:
		if r.SourceThreadID == "" {
			return relocationNack("wrong_method_thread_fork")
		}
	case RelocationMethodThreadStart:
		if r.SourceThreadID != "" {
			return relocationNack("wrong_method_thread_start")
		}
	case "SessionStart":
		return relocationNack("session_start_not_evidence")
	case "turn/start":
		return relocationNack("empty_turn_not_evidence")
	case "turn/steer":
		return relocationNack("turn_steer_not_evidence")
	default:
		return relocationNack("unknown_method")
	}
	switch {
	case !r.ThreadStarted:
		return relocationNack("thread_started_not_observed")
	case r.ThreadID == "":
		return relocationNack("empty_thread_id")
	case r.ThreadID == r.SourceThreadID:
		return relocationNack("thread_id_not_new")
	case r.ForkedFromID != r.SourceThreadID:
		// Fork: lineage must name the source. Start: lineage must be empty.
		return relocationNack("forked_from_mismatch")
	case r.RequestCwd != h.TargetPath || r.ResponseCwd != h.TargetPath || r.ReadbackCwd != h.TargetPath:
		return relocationNack("cwd_mismatch")
	case r.ReadbackBranch != h.TargetBranch:
		return relocationNack("branch_mismatch")
	case r.ReadbackHead != h.DevelopPin:
		return relocationNack("head_mismatch")
	}
	return nil
}

// RecordHeadlessRelocation stores validated relocation evidence for a
// SWITCH_PENDING_HEADLESS handoff, once. The state, mode, and nonce are
// re-read inside the write transaction; the handoff stays SWITCH_PENDING.
func (s *Store) RecordHeadlessRelocation(ctx context.Context, h Handoff, r HeadlessRelocation) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var stored Handoff
	if err := tx.QueryRowContext(ctx, `SELECT state,mode,target_path,target_branch,develop_pin FROM lane_handoffs WHERE id=? AND nonce=?`, h.ID, h.Nonce).
		Scan(&stored.State, &stored.Mode, &stored.TargetPath, &stored.TargetBranch, &stored.DevelopPin); err != nil {
		return fmt.Errorf("handoff %s: %w", h.ID, err)
	}
	if stored.State != HandoffSwitchPendingHeadless || stored.Mode != HandoffModeHeadless {
		return fmt.Errorf("handoff %s is %s/%s, not %s", h.ID, stored.Mode, stored.State, HandoffSwitchPendingHeadless)
	}
	if err := validateRelocation(stored, r); err != nil {
		return err
	}
	started := 0
	if r.ThreadStarted {
		started = 1
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO lane_handoff_relocations(handoff_id,nonce,method,source_thread_id,thread_id,forked_from_id,thread_started,request_cwd,response_cwd,readback_cwd,readback_branch,readback_head,recorded_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		h.ID, h.Nonce, r.Method, r.SourceThreadID, r.ThreadID, r.ForkedFromID, started, r.RequestCwd, r.ResponseCwd,
		r.ReadbackCwd, r.ReadbackBranch, r.ReadbackHead, s.now().UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	return tx.Commit()
}

// HeadlessRelocationFor reads the relocation evidence recorded for a handoff.
func (s *Store) HeadlessRelocationFor(ctx context.Context, handoffID string) (HeadlessRelocation, bool, error) {
	var r HeadlessRelocation
	var started int
	var at string
	err := s.db.QueryRowContext(ctx, `SELECT handoff_id,nonce,method,source_thread_id,thread_id,forked_from_id,thread_started,request_cwd,response_cwd,readback_cwd,readback_branch,readback_head,recorded_at FROM lane_handoff_relocations WHERE handoff_id=?`, handoffID).
		Scan(&r.HandoffID, &r.Nonce, &r.Method, &r.SourceThreadID, &r.ThreadID, &r.ForkedFromID, &started, &r.RequestCwd, &r.ResponseCwd,
			&r.ReadbackCwd, &r.ReadbackBranch, &r.ReadbackHead, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return HeadlessRelocation{}, false, nil
	}
	if err != nil {
		return HeadlessRelocation{}, false, err
	}
	r.ThreadStarted = started == 1
	r.RecordedAt, _ = time.Parse(time.RFC3339Nano, at)
	return r, true, nil
}
