package cli

// factory_lane_handoff_recover.go — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 M4.
//
// Restart recovery of a lane worktree handoff (REQ-FLH-011). After a crash at
// any point — reservation, creation, branch rename, SWITCH_PENDING, the rebind
// transaction, or receipt delivery — the restarted controller rereads the
// broker, filesystem, and Git facts and selects exactly one of resume,
// idempotent finalize, NACK, or ABANDONED. It never deletes a worktree and
// never guesses past a fact it cannot prove.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Named crash points of the handoff controller. laneHandoffFailpoint runs at
// each; production never sets it.
const (
	handoffPointReserved      = "reserved"       // reservation committed, target not created
	handoffPointCreated       = "created"        // materializer ran, branch not renamed
	handoffPointRenamed       = "renamed"        // branch renamed, provenance not verified
	handoffPointSwitchPending = "switch-pending" // SWITCH_PENDING committed (headless: before the RPC)
	handoffPointRelocated     = "relocated"      // headless relocation recorded, rebind not run
	handoffPointBound         = "bound"          // rebind committed, receipt not delivered to the caller
)

// laneHandoffFailpoint is a test-only fault-injection seam: a test makes it
// panic at one crash point to model the controller dying there.
var laneHandoffFailpoint = func(string) {}

// Recovery decisions.
const (
	laneRecoveryNone      = "none"      // no non-final handoff: nothing to recover
	laneRecoveryResume    = "resume"    // continued (or already sits) at a safe state
	laneRecoveryFinalize  = "finalize"  // BOUND: the durable receipt is read back, nothing written
	laneRecoveryNack      = "nack"      // provably unfinishable: NACK, fresh reservation only
	laneRecoveryAbandoned = "abandoned" // not provably safe: ABANDONED, target preserved
)

// laneHandoffRecoverRequest names the lane to recover and the process that
// owns a relocated headless thread's endpoint (for a resumed rebind).
type laneHandoffRecoverRequest struct {
	ProjectRoot, RunID, Slot string
	Owner                    laneHandoffBind
}

// laneHandoffRecovery is the decision a restart made for the lane's newest
// handoff, with the NACK/ABANDONED reason or the finalized receipt.
type laneHandoffRecovery struct {
	Decision, Reason string
	Handoff          factorymsg.Handoff
	Binding          *factorymsg.HandoffBinding
}

// recoverLaneHandoff is the restart reconciler. It reads the lane's newest
// handoff and decides from facts alone:
//
//   - NACK/ABANDONED or no handoff: none.
//   - BOUND: finalize — read the durable receipt, tombstone, and release back
//     and write nothing; a BOUND row missing any of them is corruption and is
//     reported, never repaired by guessing.
//   - non-final with a target that is not provably ours and clean at the pin
//     (unregistered, dirty, unmerged commits, drifted base, other branch):
//     ABANDONED with that reason, target preserved.
//   - RESERVED: resume creation (materialize if absent, rename if the crash
//     came before the rename), verify provenance, and mark WT_READY.
//   - WT_READY, SWITCH_PENDING_INTERACTIVE: resume — the state is already safe.
//   - SWITCH_PENDING_HEADLESS: resume the rebind from the recorded official
//     relocation; with none recorded the RPC result is lost, so NACK.
//
// @MX:ANCHOR: [AUTO] single restart-recovery decision point for a lane handoff
// @MX:REASON: REQ-FLH-011 / AC-FLH-009/010 — every crash point must land on exactly one of resume, finalize, NACK, ABANDONED; a second recovery path could delete or bind what this one preserves
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func recoverLaneHandoff(ctx context.Context, req laneHandoffRecoverRequest, deps laneHandoffDeps) (laneHandoffRecovery, error) {
	if deps.Materialize == nil {
		deps.Materialize = materializeSessionWorktree
	}
	if deps.Out == nil {
		deps.Out = io.Discard
	}
	primary := homestate.CanonicalProjectRoot(req.ProjectRoot)
	runID, err := factorymsg.ResolveActiveRun(ctx, primary, req.RunID)
	if err != nil {
		return laneHandoffRecovery{}, err
	}
	store, err := factorymsg.Open(primary, runID)
	if err != nil {
		return laneHandoffRecovery{}, err
	}
	defer func() { _ = store.Close() }()
	hs, err := store.HandoffsForLane(ctx, req.Slot)
	if err != nil {
		return laneHandoffRecovery{}, err
	}
	if len(hs) == 0 {
		return laneHandoffRecovery{Decision: laneRecoveryNone}, nil
	}
	h := hs[len(hs)-1]
	out := laneHandoffRecovery{Handoff: h}
	switch h.State {
	case factorymsg.HandoffNack, factorymsg.HandoffAbandoned:
		out.Decision = laneRecoveryNone
		return out, nil
	case factorymsg.HandoffBound:
		b, ok, err := store.HandoffReceipt(ctx, h.ID)
		if err != nil {
			return out, err
		}
		if !ok {
			return out, fmt.Errorf("handoff %s is BOUND without its receipt, release, and tombstone: refusing to guess", h.ID)
		}
		out.Decision, out.Binding = laneRecoveryFinalize, &b
		return out, nil
	}

	settle := func(decision, reason string) (laneHandoffRecovery, error) {
		var err error
		switch decision {
		case laneRecoveryAbandoned:
			h, err = store.AbandonHandoff(ctx, h, reason)
		case laneRecoveryNack:
			h, err = store.NackHandoff(ctx, h, reason)
		}
		out.Decision, out.Reason, out.Handoff = decision, reason, h
		return out, err
	}
	if _, err := os.Lstat(h.TargetPath); err != nil {
		if !os.IsNotExist(err) || h.State != factorymsg.HandoffReserved {
			return settle(laneRecoveryAbandoned, factorymsg.NackTargetMissing)
		}
		// The crash came before creation: create now, exactly as prepare would.
		if reason := createHandoffTarget(h, primary, deps); reason != "" {
			return settle(laneRecoveryNack, reason)
		}
	} else if reason := recoverableTarget(h, primary); reason != "" {
		return settle(laneRecoveryAbandoned, reason)
	}

	switch h.State {
	case factorymsg.HandoffReserved:
		if reason := verifyCreatedTarget(h, primary); reason != "" {
			return settle(laneRecoveryAbandoned, reason)
		}
		h, err = store.MarkHandoffWTReady(ctx, h)
		out.Decision, out.Handoff = laneRecoveryResume, h
		return out, err
	case factorymsg.HandoffSwitchPendingHeadless:
		if _, ok, err := store.HeadlessRelocationFor(ctx, h.ID); err != nil {
			return out, err
		} else if !ok {
			return settle(laneRecoveryNack, factorymsg.NackRelocationEvidenceInvalid)
		}
		owner := req.Owner
		owner.ProjectRoot = primary
		b, err := bindLaneHandoffHeadless(ctx, h, owner)
		var nack *factorymsg.HandoffNackError
		if errors.As(err, &nack) {
			// The rebind recorded its own NACK (evidence or readback mismatch).
			out.Decision, out.Reason = laneRecoveryNack, nack.Reason
			out.Handoff.State, out.Handoff.Reason = factorymsg.HandoffNack, nack.Reason
			return out, nil
		}
		if err != nil {
			return out, err
		}
		out.Decision, out.Binding = laneRecoveryResume, &b
		out.Handoff.State = factorymsg.HandoffBound
		return out, nil
	}
	out.Decision = laneRecoveryResume
	return out, nil
}

// recoverableTarget reads an existing target back and returns the ABANDONED
// reason when it cannot be proven to be this handoff's clean, unadvanced card
// worktree; "" means creation may continue from it. A branch still named by
// the card id is the crash between creation and rename, and is renamed here.
func recoverableTarget(h factorymsg.Handoff, primary string) string {
	list, err := handoffGitOutput(primary, "worktree", "list", "--porcelain")
	if err != nil || !strings.Contains(list+"\n", "worktree "+h.TargetPath+"\n") {
		return factorymsg.NackOwnerUnknown
	}
	if out, err := handoffGitOutput(h.TargetPath, "status", "--porcelain"); err != nil || out != "" {
		return factorymsg.NackTargetDirty
	}
	branch, err := handoffGitOutput(h.TargetPath, "branch", "--show-current")
	if err != nil {
		return factorymsg.NackBranchCollision
	}
	if branch == h.CardID && h.State == factorymsg.HandoffReserved {
		if _, err := handoffGitOutput(h.TargetPath, "branch", "-m", h.CardID, h.TargetBranch); err != nil {
			return factorymsg.NackBranchCollision
		}
		branch = h.TargetBranch
	}
	if branch != h.TargetBranch {
		return factorymsg.NackBranchCollision
	}
	head, err := handoffGitOutput(h.TargetPath, "rev-parse", "HEAD")
	if err != nil {
		return factorymsg.NackBaseDrift
	}
	if head != h.DevelopPin {
		if handoffCommand("git", "-C", h.TargetPath, "merge-base", "--is-ancestor", h.DevelopPin, head).Run() == nil {
			return factorymsg.NackTargetUnmerged
		}
		return factorymsg.NackBaseDrift
	}
	return ""
}
