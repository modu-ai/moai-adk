package cli

// factory_lane_handoff.go — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 M1.
//
// Preparation of a stable lane's move into a card-dedicated L1 worktree:
// fail-closed admission, local develop pin, a durable broker reservation, the
// existing MoAI materializer, and exact provenance verification up to
// WT_READY. Relocation (SWITCH_PENDING_*) and the atomic rebind are M2/M3.

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// laneHandoffActivity is the mode adapter's observation of the source lane.
type laneHandoffActivity string

const (
	laneActivityIdle           laneHandoffActivity = "idle"
	laneActivityActiveTurn     laneHandoffActivity = "active_turn"
	laneActivityPermissionWait laneHandoffActivity = "permission_wait"
	laneActivityInterrupting   laneHandoffActivity = "interrupting"
)

// handoffSlug follows the WT-<slug> rule: at most 3 lowercase tokens, 24 chars.
var handoffSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+){0,2}$`)

type laneHandoffRequest struct {
	ProjectRoot, RunID               string
	Slot, CardID, SpecID, Slug, Mode string
	SourceCwd                        string
	Activity                         laneHandoffActivity
	ExpectedSource                   *factorymsg.Peer
}

// laneHandoffAppServer is the headless relocation client. Preparation never
// calls it; the M2 headless switch step does.
type laneHandoffAppServer interface {
	ForkThread(ctx context.Context, sourceThreadID, cwd string) (string, error)
	StartThread(ctx context.Context, cwd string) (string, error)
}

type laneHandoffDeps struct {
	Materialize func(name string, out io.Writer) (string, error)
	AppServer   laneHandoffAppServer
	Out         io.Writer
}

// laneHandoffTrace is the traceability envelope a WT_READY handoff carries
// into dispatch: card and SPEC ids, the card worktree and branch, and the
// card-scoped evidence path. Commits on the branch name the card id.
type laneHandoffTrace struct {
	Card, Spec, Worktree, Branch, Evidence string
}

func laneHandoffTraceOf(h factorymsg.Handoff) laneHandoffTrace {
	return laneHandoffTrace{
		Card: h.CardID, Spec: h.SpecID, Worktree: h.TargetPath, Branch: h.TargetBranch,
		Evidence: filepath.ToSlash(filepath.Join(".moai", "reports", h.CardID, "verdict.md")),
	}
}

func handoffGitOutput(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	return strings.TrimSpace(string(out)), err
}

func handoffRefExists(dir, ref string) bool {
	return exec.Command("git", "-C", dir, "show-ref", "--verify", "--quiet", ref).Run() == nil
}

func canonicalPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

func pathWithin(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func activityNack(a laneHandoffActivity) error {
	switch a {
	case laneActivityIdle:
		return nil
	case laneActivityActiveTurn:
		return factorymsg.NewHandoffNack(factorymsg.NackLaneActiveTurn, "")
	case laneActivityPermissionWait:
		return factorymsg.NewHandoffNack(factorymsg.NackLanePermissionWait, "")
	case laneActivityInterrupting:
		return factorymsg.NewHandoffNack(factorymsg.NackLaneInterrupting, "")
	}
	return factorymsg.NewHandoffNack(factorymsg.NackInvalidRequest, "unknown lane activity")
}

// prepareLaneHandoff admits and reserves a lane handoff, then materializes and
// verifies its card worktree. It returns a WT_READY handoff, or a NACK error.
// Admission NACKs write nothing; a NACK after the reservation is recorded on
// the handoff and the target is preserved for inspection, never removed.
//
// @MX:ANCHOR: [AUTO] single preparation entry of the lane worktree handoff (M1)
// @MX:REASON: REQ-FLH-003/004/005/016 — admission order, develop pin, and provenance checks are the only road to WT_READY; M2 relocation starts from its result
// @MX:SPEC: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
func prepareLaneHandoff(ctx context.Context, req laneHandoffRequest, deps laneHandoffDeps) (factorymsg.Handoff, error) {
	if err := activityNack(req.Activity); err != nil {
		return factorymsg.Handoff{}, err
	}
	if !handoffSlug.MatchString(req.Slug) || len(req.Slug) > 24 {
		return factorymsg.Handoff{}, factorymsg.NewHandoffNack(factorymsg.NackInvalidRequest, "invalid branch slug")
	}
	if deps.Materialize == nil {
		deps.Materialize = materializeSessionWorktree
	}
	if deps.Out == nil {
		deps.Out = io.Discard
	}
	primary := homestate.CanonicalProjectRoot(req.ProjectRoot)
	runID, err := factorymsg.ResolveActiveRun(ctx, primary, req.RunID)
	if err != nil {
		return factorymsg.Handoff{}, err
	}
	target := filepath.Join(primary, ".claude", "worktrees", req.CardID)
	branch := "WT-" + req.Slug

	// Cwd is untrusted input: it must resolve inside the canonical primary.
	source, err := canonicalPath(req.SourceCwd)
	if err != nil || !pathWithin(source, primary) {
		return factorymsg.Handoff{}, factorymsg.NewHandoffNack(factorymsg.NackUntrustedCwd, req.SourceCwd)
	}

	store, err := factorymsg.Open(primary, runID)
	if err != nil {
		return factorymsg.Handoff{}, err
	}
	defer func() { _ = store.Close() }()

	reentry, err := isHandoffReentry(ctx, store, req, source, target)
	if err != nil {
		return factorymsg.Handoff{}, err
	}
	if !reentry {
		if out, err := handoffGitOutput(source, "status", "--porcelain"); err != nil || out != "" {
			return factorymsg.Handoff{}, factorymsg.NewHandoffNack(factorymsg.NackSourceDirty, source)
		}
		if _, err := os.Lstat(target); err == nil {
			return factorymsg.Handoff{}, factorymsg.NewHandoffNack(factorymsg.NackTargetPathConflict, target)
		}
		if handoffRefExists(primary, "refs/heads/"+branch) || handoffRefExists(primary, "refs/heads/"+req.CardID) {
			return factorymsg.Handoff{}, factorymsg.NewHandoffNack(factorymsg.NackBranchCollision, branch)
		}
	}
	pin, err := handoffGitOutput(primary, "rev-parse", "--verify", "refs/heads/develop^{commit}")
	if err != nil {
		return factorymsg.Handoff{}, fmt.Errorf("read local develop pin: %w", err)
	}

	h, err := store.ReserveHandoff(ctx, factorymsg.HandoffReservation{
		Slot: req.Slot, CardID: req.CardID, SpecID: req.SpecID, Mode: req.Mode,
		DevelopPin: pin, TargetPath: target, TargetBranch: branch, ExpectedSource: req.ExpectedSource,
	})
	if err != nil {
		return factorymsg.Handoff{}, err
	}
	var reason string
	if reentry {
		reason = verifyReentryTarget(h)
	} else {
		reason = createHandoffTarget(h, primary, deps)
	}
	if reason != "" {
		if _, err := store.NackHandoff(ctx, h, reason); err != nil {
			return factorymsg.Handoff{}, err
		}
		return factorymsg.Handoff{}, factorymsg.NewHandoffNack(reason, h.TargetPath)
	}
	return store.MarkHandoffWTReady(ctx, h)
}

// isHandoffReentry reports the N6/N10 re-entry: the lane already sits in the
// card's target after that card's previous handoff was NACKed. That cwd and
// target are then neither untrusted nor conflicting (REQ-FLH-003).
func isHandoffReentry(ctx context.Context, store *factorymsg.Store, req laneHandoffRequest, source, target string) (bool, error) {
	if source != target {
		return false, nil
	}
	hs, err := store.HandoffsForLane(ctx, req.Slot)
	if err != nil {
		return false, err
	}
	for i := len(hs) - 1; i >= 0; i-- {
		if hs[i].CardID == req.CardID {
			return hs[i].State == factorymsg.HandoffNack && hs[i].TargetPath == target, nil
		}
	}
	return false, nil
}

// verifyReentryTarget reconstructs WT_READY without the materializer, checking
// in the REQ-FLH-003 order: dirty, branch, then pin.
func verifyReentryTarget(h factorymsg.Handoff) string {
	if out, err := handoffGitOutput(h.TargetPath, "status", "--porcelain"); err != nil || out != "" {
		return factorymsg.NackTargetDirty
	}
	if b, err := handoffGitOutput(h.TargetPath, "branch", "--show-current"); err != nil || b != h.TargetBranch {
		return factorymsg.NackBranchCollision
	}
	if head, err := handoffGitOutput(h.TargetPath, "rev-parse", "HEAD"); err != nil || head != h.DevelopPin {
		return factorymsg.NackBaseDrift
	}
	return ""
}

// createHandoffTarget runs the existing materializer once, renames its branch
// to the reserved WT-<slug>, and verifies exact provenance. Creation-base drift
// is never corrected by fast-forward or merge.
func createHandoffTarget(h factorymsg.Handoff, primary string, deps laneHandoffDeps) string {
	created, err := deps.Materialize(h.CardID, deps.Out)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Out, "moai: handoff worktree creation failed: %v\n", err)
		return factorymsg.NackTargetPathConflict
	}
	if got, err := canonicalPath(created); err != nil || got != h.TargetPath {
		return factorymsg.NackTargetPathConflict
	}
	if _, err := handoffGitOutput(h.TargetPath, "branch", "-m", h.CardID, h.TargetBranch); err != nil {
		return factorymsg.NackBranchCollision
	}
	head, err := handoffGitOutput(h.TargetPath, "rev-parse", "HEAD")
	if err != nil || head != h.DevelopPin {
		return factorymsg.NackBaseDrift
	}
	if develop, err := handoffGitOutput(primary, "rev-parse", "--verify", "refs/heads/develop^{commit}"); err != nil || develop != h.DevelopPin {
		return factorymsg.NackBaseDrift
	}
	if b, err := handoffGitOutput(h.TargetPath, "branch", "--show-current"); err != nil || b != h.TargetBranch {
		return factorymsg.NackBranchCollision
	}
	list, err := handoffGitOutput(primary, "worktree", "list", "--porcelain")
	if err != nil || strings.Count(list+"\n", "branch refs/heads/"+h.TargetBranch+"\n") != 1 {
		return factorymsg.NackBranchCollision
	}
	if out, err := handoffGitOutput(h.TargetPath, "status", "--porcelain"); err != nil || out != "" {
		return factorymsg.NackTargetDirty
	}
	return ""
}
