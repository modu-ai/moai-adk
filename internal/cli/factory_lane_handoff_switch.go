package cli

// factory_lane_handoff_switch.go — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 M2.
//
// The two relocation adapters that move a WT_READY handoff to its
// mode-specific SWITCH_PENDING state. Interactive emits operator guidance for
// a user-executed /cd and nothing else; headless relocates through the
// official app-server (thread/fork or thread/start) and records the returned
// evidence plus the controller's own target readback. Neither adapter writes
// BOUND, a tombstone, a peer change, or a dispatch release — that is the M3
// atomic rebind.

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// laneHandoffSwitch is one relocation attempt. Activity is the adapter's
// observation of the source lane at switch time; SourceThreadID names the
// headless source thread whose stored history is forked ("" = no history).
type laneHandoffSwitch struct {
	ProjectRoot    string
	Activity       laneHandoffActivity
	SourceThreadID string
}

// codexLaneHandoffAppServer is the production headless relocation client: the
// existing codex app-server client, bounded by its own deadline.
type codexLaneHandoffAppServer struct{}

func (codexLaneHandoffAppServer) ForkThread(ctx context.Context, sourceThreadID, cwd string) (codexThreadRelocation, error) {
	return relocateViaCodex(ctx, sourceThreadID, cwd)
}

func (codexLaneHandoffAppServer) StartThread(ctx context.Context, cwd string) (codexThreadRelocation, error) {
	return relocateViaCodex(ctx, "", cwd)
}

func relocateViaCodex(ctx context.Context, sourceThreadID, cwd string) (codexThreadRelocation, error) {
	bin, err := codexLookPath(codexBinaryName)
	if err != nil {
		return codexThreadRelocation{}, fmt.Errorf("codex binary: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, config.DefaultCodexHandoffRelocationTimeout)
	defer cancel()
	return runCodexThreadRelocation(ctx, bin, sourceThreadID, cwd)
}

// laneHandoffReadback is the controller's direct read of the target worktree.
type laneHandoffReadback struct{ Cwd, Branch, Head string }

func readbackHandoffTarget(target string) (laneHandoffReadback, error) {
	top, err := handoffGitOutput(target, "rev-parse", "--show-toplevel")
	if err != nil {
		return laneHandoffReadback{}, err
	}
	cwd, err := canonicalPath(top)
	if err != nil {
		return laneHandoffReadback{}, err
	}
	branch, err := handoffGitOutput(target, "branch", "--show-current")
	if err != nil {
		return laneHandoffReadback{}, err
	}
	head, err := handoffGitOutput(target, "rev-parse", "HEAD")
	if err != nil {
		return laneHandoffReadback{}, err
	}
	return laneHandoffReadback{Cwd: cwd, Branch: branch, Head: head}, nil
}

func (r laneHandoffReadback) matches(h factorymsg.Handoff) bool {
	return r.Cwd == h.TargetPath && r.Branch == h.TargetBranch && r.Head == h.DevelopPin
}

func openHandoffStore(sw laneHandoffSwitch, h factorymsg.Handoff) (*factorymsg.Store, error) {
	return factorymsg.Open(homestate.CanonicalProjectRoot(sw.ProjectRoot), h.RunID)
}

// nackSwitch records a NACK error on the handoff and returns that error. An
// error that is not a NACK is returned unchanged and records nothing.
func nackSwitch(ctx context.Context, store *factorymsg.Store, h factorymsg.Handoff, cause error) error {
	var nack *factorymsg.HandoffNackError
	if !errors.As(cause, &nack) {
		return cause
	}
	if _, err := store.NackHandoff(ctx, h, nack.Reason); err != nil {
		return err
	}
	return cause
}

// switchLaneHandoffInteractive moves an interactive WT_READY handoff to
// SWITCH_PENDING_INTERACTIVE on a lane confirmed idle, then tells the operator
// to run /cd themselves. It never reaches the app-server and never creates a
// model turn: the binding evidence is the user's next normal turn (M3).
//
// @MX:NOTE: [AUTO] interactive relocation is guidance only — the /cd is the user's act, never automated
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func switchLaneHandoffInteractive(ctx context.Context, h factorymsg.Handoff, sw laneHandoffSwitch, deps laneHandoffDeps) (factorymsg.Handoff, error) {
	if h.Mode != factorymsg.HandoffModeInteractive {
		return factorymsg.Handoff{}, fmt.Errorf("handoff %s is %s, not interactive", h.ID, h.Mode)
	}
	store, err := openHandoffStore(sw, h)
	if err != nil {
		return factorymsg.Handoff{}, err
	}
	defer func() { _ = store.Close() }()
	if err := activityNack(sw.Activity); err != nil {
		return factorymsg.Handoff{}, nackSwitch(ctx, store, h, err)
	}
	pending, err := store.MarkHandoffSwitchPendingInteractive(ctx, h)
	if err != nil {
		return factorymsg.Handoff{}, err
	}
	out := deps.Out
	if out == nil {
		out = io.Discard
	}
	_, _ = fmt.Fprintf(out, "moai: lane %s is ready to move into card %s (%s).\n"+
		"Run this yourself in that lane's Codex prompt, then send your next normal message:\n"+
		"/cd %s\n"+
		"handoff nonce: %s\n"+
		"The lane edits no code until the handoff is BOUND.\n",
		pending.Slot, pending.CardID, pending.SpecID, pending.TargetPath, pending.Nonce)
	return pending, nil
}

// switchLaneHandoffHeadless relocates an idle headless lane through the
// official app-server: thread/fork(cwd) when the source thread has stored
// history, thread/start(cwd) when it has none. It records the returned thread
// id, lineage, observed thread/started, and the controller's target readback
// as evidence for the M3 rebind, leaving the handoff SWITCH_PENDING_HEADLESS.
//
// @MX:NOTE: [AUTO] headless relocation evidence is the official RPC result plus controller readback — never SessionStart or a model turn
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func switchLaneHandoffHeadless(ctx context.Context, h factorymsg.Handoff, sw laneHandoffSwitch, deps laneHandoffDeps) (factorymsg.Handoff, error) {
	if h.Mode != factorymsg.HandoffModeHeadless {
		return factorymsg.Handoff{}, fmt.Errorf("handoff %s is %s, not headless", h.ID, h.Mode)
	}
	store, err := openHandoffStore(sw, h)
	if err != nil {
		return factorymsg.Handoff{}, err
	}
	defer func() { _ = store.Close() }()
	if err := activityNack(sw.Activity); err != nil {
		return factorymsg.Handoff{}, nackSwitch(ctx, store, h, err)
	}
	// Never ask the app-server to relocate into a target that no longer reads
	// back as reserved.
	if rb, err := readbackHandoffTarget(h.TargetPath); err != nil || !rb.matches(h) {
		return factorymsg.Handoff{}, nackSwitch(ctx, store, h, factorymsg.NewHandoffNack(factorymsg.NackTargetReadbackMismatch, h.TargetPath))
	}
	app := deps.AppServer
	if app == nil {
		app = codexLaneHandoffAppServer{}
	}
	pending, err := store.MarkHandoffSwitchPendingHeadless(ctx, h)
	if err != nil {
		return factorymsg.Handoff{}, err
	}
	var rel codexThreadRelocation
	if sw.SourceThreadID != "" {
		rel, err = app.ForkThread(ctx, sw.SourceThreadID, pending.TargetPath)
	} else {
		rel, err = app.StartThread(ctx, pending.TargetPath)
	}
	if err != nil {
		return factorymsg.Handoff{}, nackSwitch(ctx, store, pending, factorymsg.NewHandoffNack(factorymsg.NackRelocationRPCFailed, err.Error()))
	}
	rb, err := readbackHandoffTarget(pending.TargetPath)
	if err != nil {
		return factorymsg.Handoff{}, nackSwitch(ctx, store, pending, factorymsg.NewHandoffNack(factorymsg.NackTargetReadbackMismatch, err.Error()))
	}
	evidence := factorymsg.HeadlessRelocation{
		Method: rel.Method, SourceThreadID: rel.SourceThreadID, ThreadID: rel.ThreadID, ForkedFromID: rel.ForkedFromID,
		ThreadStarted: rel.ThreadStarted, RequestCwd: rel.RequestCwd, ResponseCwd: rel.ResponseCwd,
		ReadbackCwd: rb.Cwd, ReadbackBranch: rb.Branch, ReadbackHead: rb.Head,
	}
	if err := store.RecordHeadlessRelocation(ctx, pending, evidence); err != nil {
		return factorymsg.Handoff{}, nackSwitch(ctx, store, pending, err)
	}
	return pending, nil
}
