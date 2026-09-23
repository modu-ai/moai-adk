package cli

// factory_lane_handoff_bind.go — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 M3.
//
// The headless half of the atomic rebind: bind a SWITCH_PENDING_HEADLESS
// handoff directly from the official relocation result M2 recorded (the
// returned thread id) plus the controller's fresh readback of the target. No
// hook event, no model turn, and no second app-server request is involved.
// The interactive half lives in the hook (the user's next-turn evidence).

import (
	"context"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// laneHandoffBind names the process that owns the relocated headless thread.
type laneHandoffBind struct {
	ProjectRoot       string
	OwnerPID          int
	OwnerProcessStart string
}

// bindLaneHandoffHeadless reads the recorded relocation and the target back,
// then runs the broker's single rebind transaction. A missing relocation or a
// target that no longer reads back as reserved NACKs the handoff; the rebind
// itself NACKs evidence it rejects and refuses a stale reservation.
//
// @MX:NOTE: [AUTO] headless BOUND comes from the recorded thread id plus a fresh controller readback — never from a hook event
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func bindLaneHandoffHeadless(ctx context.Context, h factorymsg.Handoff, b laneHandoffBind) (factorymsg.HandoffBinding, error) {
	if h.Mode != factorymsg.HandoffModeHeadless {
		return factorymsg.HandoffBinding{}, fmt.Errorf("handoff %s is %s, not headless", h.ID, h.Mode)
	}
	store, err := factorymsg.Open(homestate.CanonicalProjectRoot(b.ProjectRoot), h.RunID)
	if err != nil {
		return factorymsg.HandoffBinding{}, err
	}
	defer func() { _ = store.Close() }()
	rel, ok, err := store.HeadlessRelocationFor(ctx, h.ID)
	if err != nil {
		return factorymsg.HandoffBinding{}, err
	}
	if !ok {
		return factorymsg.HandoffBinding{}, nackSwitch(ctx, store, h, factorymsg.NewHandoffNack(factorymsg.NackRelocationEvidenceInvalid, "no recorded relocation"))
	}
	rb, err := readbackHandoffTarget(h.TargetPath)
	if err != nil {
		return factorymsg.HandoffBinding{}, nackSwitch(ctx, store, h, factorymsg.NewHandoffNack(factorymsg.NackTargetReadbackMismatch, err.Error()))
	}
	binding, err := store.BindHandoff(ctx, h, factorymsg.HandoffBindEvidence{
		Mode: factorymsg.HandoffModeHeadless, Nonce: h.Nonce, CardID: h.CardID, SpecID: h.SpecID,
		SessionUUID: rel.ThreadID, PID: b.OwnerPID, ProcessStart: b.OwnerProcessStart,
		Cwd: rb.Cwd, WorktreeRoot: rb.Cwd, Branch: rb.Branch, Head: rb.Head,
	})
	if err == nil {
		laneHandoffFailpoint(handoffPointBound)
	}
	return binding, err
}
