package hook

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

// factoryWorktreeReadback is the hook's own Git read of the SessionStart cwd:
// the canonical cwd, its worktree root, branch, and HEAD. A field that cannot
// be read stays empty and fails the rebind's comparison.
type factoryWorktreeReadback struct{ Cwd, Root, Branch, Head string }

func readbackFactoryWorktree(ctx context.Context, cwd string) factoryWorktreeReadback {
	var rb factoryWorktreeReadback
	if c, err := filepath.EvalSymlinks(cwd); err == nil {
		rb.Cwd = c
	}
	git := func(args ...string) string {
		out, err := exec.CommandContext(ctx, "git", append([]string{"-C", cwd}, args...)...).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	if top := git("rev-parse", "--show-toplevel"); top != "" {
		if c, err := filepath.EvalSymlinks(top); err == nil {
			rb.Root = c
		}
	}
	rb.Branch = git("branch", "--show-current")
	rb.Head = git("rev-parse", "HEAD")
	return rb
}

// bindFactoryInteractiveHandoff consumes the SessionStart of the user's next
// normal turn after /cd as interactive binding evidence (REQ-FLH-006/008). It
// acts only on the lane's SWITCH_PENDING_INTERACTIVE handoff and only for a
// session that is not the lane's current endpoint (the caller has already
// returned for a current session). Everything else — no handoff, a WT_READY
// or headless handoff — is not handled here and follows t1074 semantics.
//
// @MX:NOTE: [AUTO] interactive BOUND is triggered only by SessionStart evidence; UserPromptSubmit never binds a handoff
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func bindFactoryInteractiveHandoff(ctx context.Context, s *factorymsg.Store, input *HookInput, want factorymsg.Peer) (string, bool) {
	hs, err := s.HandoffsForLane(ctx, want.Slot)
	if err != nil {
		return "factory messaging degraded: " + err.Error(), true
	}
	var h factorymsg.Handoff
	for i := len(hs) - 1; i >= 0; i-- {
		if hs[i].State == factorymsg.HandoffSwitchPendingInteractive {
			h = hs[i]
			break
		}
	}
	if h.ID == "" {
		return "", false
	}
	if input.CWD == "" {
		// No cwd, no evidence: the handoff stays pending.
		return "factory handoff pending: SessionStart carried no cwd", true
	}
	rb := readbackFactoryWorktree(ctx, input.CWD)
	b, err := s.BindHandoff(ctx, h, factorymsg.HandoffBindEvidence{
		Mode: factorymsg.HandoffModeInteractive, Nonce: h.Nonce, CardID: h.CardID, SpecID: h.SpecID,
		SessionUUID: input.SessionID, PID: want.PID, ProcessStart: want.ProcessStart,
		Cwd: rb.Cwd, WorktreeRoot: rb.Root, Branch: rb.Branch, Head: rb.Head,
	})
	if err != nil {
		if reason, ok := factorymsg.HandoffNackReason(err); ok {
			var nack *factorymsg.HandoffNackError
			if errors.As(err, &nack) {
				return fmt.Sprintf("factory handoff NACK %s: card=%s slot=%s; the lane keeps its previous endpoint and edits no code", reason, h.CardID, h.Slot), true
			}
			return fmt.Sprintf("factory handoff refused %s: card=%s slot=%s", reason, h.CardID, h.Slot), true
		}
		return "factory messaging degraded: " + err.Error(), true
	}
	return fmt.Sprintf("factory handoff bound: card=%s spec=%s slot=%s generation=%d receipt=%s; %d dispatch(es) released to this session",
		h.CardID, h.SpecID, b.Slot, b.New.Generation, b.ReceiptID, b.Released), true
}

// factoryHandoffRegistrationNotice names a UserPromptSubmit registration the
// broker refused because of a lane handoff (REQ-FLH-018). The endpoint was not
// rotated; the hook stays fail-open and only reports.
func factoryHandoffRegistrationNotice(err error, slot string) (string, bool) {
	if stale, ok := factorymsg.StaleEndpoint(err); ok && stale.Code == factorymsg.NackStaleEndpoint {
		return fmt.Sprintf("factory endpoint replaced %s: slot=%s is current at %s generation %d; this session receives no factory messages",
			stale.Code, slot, stale.Current.SessionUUID, stale.Current.Generation), true
	}
	if reason, ok := factorymsg.HandoffNackReason(err); ok && reason == factorymsg.NackEndpointHandoffPending {
		return fmt.Sprintf("factory handoff pending %s: slot=%s; this session is not the lane endpoint until the handoff binds, and the endpoint is unchanged", reason, slot), true
	}
	return "", false
}
