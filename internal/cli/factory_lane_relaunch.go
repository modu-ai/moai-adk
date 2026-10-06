package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/spf13/cobra"
)

// factory_lane_relaunch.go is the Claude-harness lane's supervising loop
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-020, design.md §6): the relaunch
// clear policy turns the `moai cc|glm -l --clear-policy relaunch`
// launcher into the parent that leases the next card, ensures its worktree,
// starts ONE interactive session in that worktree, waits for the operator to
// end it, and repeats — no process replacement happens on this path (the
// exec model of clear-each / clear-when-full is untouched), and the loop
// serves Windows through the same child-process form.

// factoryLaneCardLaunchFn starts one interactive child session and waits for
// it to exit. It is a package var so the tests substitute the session the
// way the codex loop's launch seam does.
var factoryLaneCardLaunchFn = func(c *exec.Cmd) error {
	return c.Run()
}

// factoryLaneRunGateFn is the lane-join gate the loop re-enters every
// iteration (SPEC-FACTORY-STALE-RUN-HEAL-001 REQ-SRH-011); factoryLaneLeaseFn
// is the lease step. Both are package vars so the tests substitute them the
// way they substitute the launch seam above.
var (
	factoryLaneRunGateFn = enterFactoryLaneRun
	factoryLaneLeaseFn   = factoryNextLeaseOnce
)

// @MX:NOTE: the loop stays the parent (design.md §6): lease the next card
// through the F1 machinery on the parent checkout, ensure its worktree,
// start ONE interactive session there, wait, repeat. The stop condition is
// `next`'s no-card answer. The child learns its card through
// config.EnvFactoryCard; the launcher process carries the lane stamps
// (enterFactoryLaneMode ran in the lane branch) and every child inherits
// them, so each fresh session re-enters the cycle with the same lane
// identity and the same clear policy.
// @MX:SPEC: SPEC-FACTORY-SELF-DISPATCH-001
//
// @MX:NOTE: the run is re-resolved on EVERY iteration (REQ-SRH-011): the
// launcher's own gate resolved it once, but a run can retire and another take
// its place while a card session runs, and a loop that kept the first id would
// keep leasing from the dead run's record. explicit and leadTarget are the
// launcher's own selectors (entry.FactoryRun / entry.FactoryLead), handed to
// the same shared gate; an explicit selection that retires stops the loop.
// @MX:SPEC: SPEC-FACTORY-STALE-RUN-HEAL-001
func runFactoryLaneRelaunch(cmd *cobra.Command, label string, claudeArgs []string, explicit, leadTarget string) error {
	// REQ-SCV-012 (SPEC-SESSION-CC-VERSION-002): derive the claude option
	// model once per process from the binary on PATH before the guard scans;
	// a derivation failure degrades silently to the snapshot and the scan
	// below stays a pure argv walk over package state.
	refreshActiveClaudeOptionModel()
	// REQ-SCV-010 (SPEC-SESSION-CC-VERSION-001): the guard runs before the
	// parent-checkout assertion and the loop's first iteration — a --resume
	// token under the relaunch policy cannot mean what it says (every card
	// session the loop starts would receive it), so the loop refuses to start
	// at all: zero leases, zero card sessions, and the token is neither
	// propagated nor stripped. The refusal names the bare one-shot lane join;
	// an operator --name beside -l is refused at the entry parse
	// (laneFlagNameError) before this guard is ever reached.
	if carriesResumeToken(claudeArgs) {
		return errors.New(relaunchResumeRefusal)
	}
	// The loop drives the F1 lease machinery itself, so it inherits the
	// `next` verb's own precondition: the parent checkout (REQ-SD-010).
	if err := factoryAssertParentCheckout(resolveProjectDir()); err != nil {
		return err
	}
	binaryPath, err := claudeLookPath(claudeBinaryName)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"factory lane: %s not found — the relaunch policy starts one %s session per card, and the CLI must be installed\n",
			claudeBinaryName, claudeBinaryName)
		return &exitCodeError{code: 1}
	}
	root := factoryCardRoot()
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		more, err := factoryLaneRelaunchIteration(ctx, cmd, root, binaryPath, claudeArgs, label, explicit, leadTarget)
		if err != nil {
			return err
		}
		if !more {
			return nil // no card available: the loop's stop condition
		}
	}
}

// factoryLaneRelaunchIteration is one pass of the relaunch loop: re-enter the
// lane-join gate, lease from the run it resolved, and run one card session.
// It reports whether a card was leased (false is the loop's stop condition).
// The gate's environment stamp is restored when the pass ends, after the
// child has exited — the child launched inside the pass inherits the stamp the
// gate just set, which is how an iteration after a run switch starts its
// session on the new run.
func factoryLaneRelaunchIteration(ctx context.Context, cmd *cobra.Command, root, binaryPath string, claudeArgs []string, label, explicit, leadTarget string) (bool, error) {
	restore, err := factoryLaneRunGateFn(launchProjectRoot(), explicit, leadTarget, nil)
	if err != nil {
		// The gate refused (no active run, an explicit run that retired,
		// ambiguous leaders): nothing is leased and the loop stops with the
		// gate's own text.
		return false, fmt.Errorf("factory lane: re-join the factory run: %w", err)
	}
	defer restore()
	runID := strings.TrimSpace(os.Getenv(config.EnvFactoryRunID))
	card, leased, err := factoryLaneLeaseFn(ctx, root, runID, label)
	if err != nil {
		return false, fmt.Errorf("factory lane: %w", err)
	}
	if !leased {
		return false, nil
	}
	wt, _, err := factoryEnsureCardWorktree(ctx, root, runID, card, label, cmd.ErrOrStderr())
	if err != nil {
		return false, fmt.Errorf("factory lane: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s stage=%s worktree=%s\n", card.CardID, dash(card.Stage), filepath.Base(wt))
	if err := launchFactoryLaneCardSession(binaryPath, claudeArgs, wt, card.CardID); err != nil {
		// On that session's exit, continue with the next card: a child
		// that failed to start or exited non-zero does not stop the
		// loop; its card stays leased until expiry.
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "factory lane: card %s session: %v\n", card.CardID, err)
	}
	return true, nil
}

// launchFactoryLaneCardSession starts ONE interactive session whose working
// directory is the card worktree (the codex per-card relaunch's shape,
// design.md D1/§6). claudeArgs is the argv the lane branch built for the
// engine — the session name the leader dispatches to plus the injected
// settings — and the child inherits the launcher environment (the lane
// stamps and the clear policy) with the leased card's id added, which is
// what makes each fresh session pick up exactly the card the launcher
// leased for it. The existing child-process launch form serves every
// platform; no syscall use on this path.
func launchFactoryLaneCardSession(binaryPath string, claudeArgs []string, wt, cardID string) error {
	env := append(os.Environ(), config.EnvFactoryCard+"="+cardID)
	c := exec.Command(binaryPath, claudeArgs...)
	c.Dir = wt
	c.Env = env
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return factoryLaneCardLaunchFn(c)
}
