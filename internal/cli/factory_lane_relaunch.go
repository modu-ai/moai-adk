package cli

import (
	"context"
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
// clear policy turns the `moai cc|glm -f lane --clear-policy relaunch`
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

// @MX:NOTE: the loop stays the parent (design.md §6): lease the next card
// through the F1 machinery on the parent checkout, ensure its worktree,
// start ONE interactive session there, wait, repeat. The stop condition is
// `next`'s no-card answer. The child learns its card through
// config.EnvMoaiKanbanCard; the launcher process carries the lane stamps
// (enterFactoryLaneMode ran in the lane branch) and every child inherits
// them, so each fresh session re-enters the cycle with the same lane
// identity and the same clear policy.
// @MX:SPEC: SPEC-FACTORY-SELF-DISPATCH-001
func runFactoryLaneRelaunch(cmd *cobra.Command, label string, claudeArgs []string) error {
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
	// The run id was resolved and stamped by the lane branch's
	// enterSelectedFactoryRun before the divert reached here.
	runID := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanID))
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		card, leased, err := factoryNextLeaseOnce(ctx, root, runID, label)
		if err != nil {
			return fmt.Errorf("factory lane: %w", err)
		}
		if !leased {
			return nil // no card available: the loop's stop condition
		}
		wt, _, err := factoryEnsureCardWorktree(ctx, root, runID, card, label, cmd.ErrOrStderr())
		if err != nil {
			return fmt.Errorf("factory lane: %w", err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s stage=%s worktree=%s\n", card.CardID, dash(card.Stage), filepath.Base(wt))
		if err := launchFactoryLaneCardSession(binaryPath, claudeArgs, wt, card.CardID); err != nil {
			// On that session's exit, continue with the next card: a child
			// that failed to start or exited non-zero does not stop the
			// loop; its card stays leased until expiry.
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "factory lane: card %s session: %v\n", card.CardID, err)
		}
	}
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
	env := append(os.Environ(), config.EnvMoaiKanbanCard+"="+cardID)
	c := exec.Command(binaryPath, claudeArgs...)
	c.Dir = wt
	c.Env = env
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return factoryLaneCardLaunchFn(c)
}
