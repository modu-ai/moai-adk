package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Terminal reasons (M4, REQ-FLH-011). SOURCE_OWNER_LIVE and
// HANDOFF_NOT_PENDING refuse the operator abandon and write nothing;
// OPERATOR_ABANDONED is the reason an accepted abandon records.
const (
	NackSourceOwnerLive   = "SOURCE_OWNER_LIVE"
	NackHandoffNotPending = "HANDOFF_NOT_PENDING"
	AbandonOperator       = "OPERATOR_ABANDONED"
)

// sourceOwnerNotCurrent reports whether the recorded source owner is provably
// not current. Live with the recorded process-start is current; Dead, or Live
// with a different known process-start (a reused PID), is not. Everything else
// — Indeterminate, an empty fingerprint — is not established, and an
// unestablished owner counts as live.
func sourceOwnerNotCurrent(probe func(int) (string, homestate.ProcessIdentityState), pid int, start string) bool {
	fp, state := probe(pid)
	switch state {
	case homestate.ProcessIdentityDead:
		return true
	case homestate.ProcessIdentityLive:
		return fp != "" && fp != start
	}
	return false
}

// AbandonLane is the operator termination of a lane stuck behind a non-final
// handoff (REQ-FLH-011). In ONE broker write transaction it rereads the lane's
// open handoff and probes its recorded source owner: an owner that is current,
// or whose liveness cannot be established, refuses with SOURCE_OWNER_LIVE; a
// lane with no open handoff refuses with HANDOFF_NOT_PENDING. Both write
// nothing. Otherwise the handoff becomes ABANDONED/OPERATOR_ABANDONED and
// nothing else is written: no BOUND, tombstone, receipt, or dispatch release,
// and the target worktree is left untouched.
//
// The probe is the three-state process identity probe, never a folded bool:
// folding Indeterminate into "not current" would abandon a live lane.
//
// @MX:WARN: [AUTO] the handoff read, the owner probe, and the ABANDONED write share one write transaction
// @MX:REASON: REQ-FLH-011 / AC-FLH-020 — a read outside the transaction could abandon a handoff a concurrent rebind already bound
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func (s *Store) AbandonLane(ctx context.Context, slot string, probe func(int) (string, homestate.ProcessIdentityState)) (Handoff, error) {
	if !safeID.MatchString(slot) {
		return Handoff{}, NewHandoffNack(NackInvalidRequest, "invalid lane")
	}
	if probe == nil {
		return Handoff{}, errors.New("abandon requires a source-owner probe")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Handoff{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var h Handoff
	err = tx.QueryRowContext(ctx, `SELECT id,run_id,card_id,spec_id,mode,nonce,state,source_session,source_generation,source_pid,source_process_start,target_path,target_branch FROM lane_handoffs WHERE slot=? AND `+openHandoffStates, slot).
		Scan(&h.ID, &h.RunID, &h.CardID, &h.SpecID, &h.Mode, &h.Nonce, &h.State, &h.Source.SessionUUID, &h.Source.Generation, &h.Source.PID, &h.Source.ProcessStart, &h.TargetPath, &h.TargetBranch)
	if errors.Is(err, sql.ErrNoRows) {
		return Handoff{}, NewHandoffNack(NackHandoffNotPending, slot)
	}
	if err != nil {
		return Handoff{}, err
	}
	h.Slot, h.Source.Slot = slot, slot
	if !sourceOwnerNotCurrent(probe, h.Source.PID, h.Source.ProcessStart) {
		return Handoff{}, NewHandoffNack(NackSourceOwnerLive, fmt.Sprintf("lane %s source owner pid %d", slot, h.Source.PID))
	}
	now := s.now().UTC()
	stamp := now.Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `UPDATE lane_handoffs SET state=?,reason=?,updated_at=? WHERE id=? AND nonce=? AND state=?`, HandoffAbandoned, AbandonOperator, stamp, h.ID, h.Nonce, h.State); err != nil {
		return Handoff{}, err
	}
	if err := insertHandoffEvent(ctx, tx, h.ID, h.State, HandoffAbandoned, AbandonOperator, stamp); err != nil {
		return Handoff{}, err
	}
	if err := tx.Commit(); err != nil {
		return Handoff{}, err
	}
	h.State, h.Reason, h.UpdatedAt = HandoffAbandoned, AbandonOperator, now
	return h, nil
}
