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

	// Restart-recovery ABANDONED reasons for a target that cannot be proven
	// safe to resume. The target is preserved in every case.
	NackTargetUnmerged = "TARGET_UNMERGED" // commits on the card branch beyond the pin
	NackOwnerUnknown   = "OWNER_UNKNOWN"   // a directory at the target this repository does not register as a worktree
	NackTargetMissing  = "TARGET_MISSING"  // a created target that no longer exists
)

// AbandonHandoff records a non-final handoff ABANDONED with reason, as a
// nonce-and-state CAS. It writes nothing else and never touches the target.
func (s *Store) AbandonHandoff(ctx context.Context, h Handoff, reason string) (Handoff, error) {
	if reason == "" {
		return Handoff{}, errors.New("handoff ABANDONED requires a reason")
	}
	return s.transitionHandoff(ctx, h, func(state, _ string) bool { return isOpenHandoffState(state) }, HandoffAbandoned, reason)
}

// HandoffReceipt reads back the durable BOUND facts of a handoff: its receipt
// and release count, and whether the replaced endpoint's tombstone exists. ok
// is false when any of the three is missing — which the one-transaction
// rebind makes impossible, so a caller treats it as corruption, not a gap to
// fill.
func (s *Store) HandoffReceipt(ctx context.Context, handoffID string) (HandoffBinding, bool, error) {
	b := HandoffBinding{HandoffID: handoffID}
	var created string
	err := s.db.QueryRowContext(ctx, `SELECT r.id,r.slot,r.old_session,r.old_generation,r.session_uuid,r.generation,r.created_at,d.message_count
		FROM lane_handoff_receipts r JOIN lane_dispatch_releases d ON d.handoff_id=r.handoff_id WHERE r.handoff_id=?`, handoffID).
		Scan(&b.ReceiptID, &b.Slot, &b.Old.SessionUUID, &b.Old.Generation, &b.New.SessionUUID, &b.New.Generation, &created, &b.Released)
	if errors.Is(err, sql.ErrNoRows) {
		return HandoffBinding{}, false, nil
	}
	if err != nil {
		return HandoffBinding{}, false, err
	}
	var tombstones int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM lane_endpoint_tombstones WHERE handoff_id=? AND session_uuid=? AND generation=?`, handoffID, b.Old.SessionUUID, b.Old.Generation).Scan(&tombstones); err != nil {
		return HandoffBinding{}, false, err
	}
	b.Old.Slot, b.New.Slot = b.Slot, b.Slot
	b.BoundAt, _ = time.Parse(time.RFC3339Nano, created)
	return b, tombstones == 1, nil
}

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
