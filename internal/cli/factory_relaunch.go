package cli

// factory_relaunch.go — the `moai factory relaunch` verb
// (SPEC-FACTORY-STALE-RUN-HEAL-001 REQ-SRH-012..014, REQ-SRH-016): the
// executable way back for a lane session whose run died or was replaced. The
// stale-run notices print its command line (internal/factory builds it from the
// same flag names registered here), and the verb re-executes the provider's own
// lane-join entry, so the shared join gate does the run resolution.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// factoryRelaunchExecFn runs the launcher child the verb re-executes. It is a
// package var so tests substitute the launch the way factoryLaneCardLaunchFn
// is substituted for the clear-policy loop.
var factoryRelaunchExecFn = func(c *exec.Cmd) error { return c.Run() }

// buildRelaunchCommand validates the verb's arguments and returns the shared
// command value. A legacy lane spelling is refused naming the canonical form
// (REQ-SRH-014, SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009: detection, never a
// mapping), and a Codex pin is refused naming the Codex limitation
// (REQ-SRH-016).
func buildRelaunchCommand(provider, lane, run, fromRun string) (factory.RelaunchCommand, error) {
	c := factory.RelaunchCommand{Provider: provider, Lane: lane, Run: run, FromRun: fromRun}
	switch provider {
	case factory.RelaunchProviderCC, factory.RelaunchProviderGLM, factory.RelaunchProviderCodex:
	default:
		return c, fmt.Errorf("--%s must be %s, %s, or %s (got %q)", factory.RelaunchFlagProvider,
			factory.RelaunchProviderCC, factory.RelaunchProviderGLM, factory.RelaunchProviderCodex, provider)
	}
	if lane != "" {
		lowered := strings.ToLower(lane)
		if factory.IsLegacyFactoryRoleValue(lowered) {
			if n, isLabel := factory.SplitFactoryLegacyLabel(lowered); isLabel {
				return c, fmt.Errorf("--%s %q is the legacy lane label; use %q", factory.RelaunchFlagLane, lane, factory.FactoryLaneLabel(n))
			}
			return c, fmt.Errorf("--%s %q is a legacy role token; use a lane-<n> label, or omit --%s to join as the next free lane",
				factory.RelaunchFlagLane, lane, factory.RelaunchFlagLane)
		}
		if _, ok := factory.SplitFactoryLaneLabel(lane); !ok {
			return c, fmt.Errorf("--%s must be a lane-<n> label (omit it to join as the next free lane), got %q", factory.RelaunchFlagLane, lane)
		}
	}
	if provider == factory.RelaunchProviderCodex {
		for _, pin := range [][2]string{{factory.RelaunchFlagLane, lane}, {factory.RelaunchFlagRun, run}} {
			if pin[1] != "" {
				return c, fmt.Errorf("--%s is not available with --%s %s: the Codex launcher accepts only '%s' (the next free lane of the single active run)",
					pin[0], factory.RelaunchFlagProvider, factory.RelaunchProviderCodex, factory.RelaunchCommand{Provider: provider}.LaunchLine())
			}
		}
	}
	for _, id := range [][2]string{{factory.RelaunchFlagRun, run}, {factory.RelaunchFlagFromRun, fromRun}} {
		if id[1] != "" && !factorymsg.ValidRunID(id[1]) {
			return c, fmt.Errorf("--%s %q is not a run id", id[0], id[1])
		}
	}
	return c, nil
}

// relaunchRetireFromRun applies REQ-SRH-013: the run named by --from-run is
// retired through the existing proof-gated predicate only when it measures
// active and its owner classifies dead. Every other case leaves the run
// untouched and returns the outcome text; the verb launches regardless.
func relaunchRetireFromRun(ctx context.Context, runID string) string {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("run %s: working directory unavailable (%v), left untouched", runID, err)
	}
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		return fmt.Sprintf("run %s: factory state unavailable (%v), left untouched", runID, err)
	}
	state, detail, err := factorymsg.ProbeRunStateAt(ctx, dbPath, runID)
	switch {
	case state == factorymsg.RunStateUnavailable:
		return fmt.Sprintf("run %s: state unavailable (%v), left untouched", runID, err)
	case state == factorymsg.RunStateNotActive && detail == "absent":
		return fmt.Sprintf("run %s: unknown (no such run), left untouched", runID)
	case state == factorymsg.RunStateNotActive:
		return fmt.Sprintf("run %s: not active (%s), left untouched", runID, detail)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Sprintf("run %s: factory state unavailable (%v), left untouched", runID, err)
	}
	defer func() { _ = db.Close() }() // the outcome is already decided; a close failure cannot change it
	classification, err := db.RetireRunIfDead(ctx, runID, factorymsg.ReconcileOptionsFor(root))
	switch {
	case err == nil:
		return fmt.Sprintf("run %s: retired (owner %s)", runID, classification)
	case errors.Is(err, homestate.ErrRunOwnerNotDead):
		return fmt.Sprintf("run %s: left active (owner %s)", runID, classification)
	default:
		return fmt.Sprintf("run %s: could not be retired (%v), left untouched", runID, err)
	}
}

// relaunchLaneIgnoredNote is the one stderr line printed when --lane is
// supplied: the launcher join the verb runs is `-l`, which takes no argument
// (SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-002), so the supplied label cannot pin
// the slot and the session joins as the next free lane.
func relaunchLaneIgnoredNote(lane string) string {
	return fmt.Sprintf("--%s %s is ignored: the launcher join -l takes no argument, so the session joins as the next free lane",
		factory.RelaunchFlagLane, lane)
}

// newFactoryRelaunchCommand builds the verb. Its flag names are the shared
// builder's constants, so a line the hook prints is accepted verbatim.
func newFactoryRelaunchCommand() *cobra.Command {
	var provider, lane, run, fromRun string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "relaunch",
		Short: "Relaunch a factory lane session into a live run",
		Long: `Relaunch a factory lane session into a live run.

The verb re-executes the provider's own lane entry — 'moai cc -l' or 'moai glm -l'
through the shared lane-join gate, 'moai codex -l' for Codex — so the run is
resolved exactly as when the operator types that line. '-l' takes no argument, so
a supplied --lane is not passed on: the verb says so on stderr and the session
joins as the next free lane.
Run it from a terminal after ending the stale session; the stale-run notice prints
the command with its arguments filled in.

--from-run <id> retires that run first, but only when it measures active and its
owner classifies dead under the run-retirement predicate; otherwise the run is left
untouched, the outcome is named, and the launch proceeds. --dry-run prints the
launch line and writes nothing.

This verb is distinct from the '--clear-policy relaunch' lane loop: that policy keeps
one launcher process looping over cards inside one run, while this verb is a one-shot
command that starts a new lane session in a live run. It never creates, edits, or
reactivates run records beyond the --from-run retirement, and it does not migrate
card leases or worktrees.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := buildRelaunchCommand(provider, lane, run, fromRun)
			if err != nil {
				return err
			}
			if dryRun {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), c.LaunchLine())
				return nil
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if fromRun != "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), relaunchRetireFromRun(ctx, fromRun))
			}
			if lane != "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), relaunchLaneIgnoredNote(lane))
			}
			self, err := os.Executable()
			if err != nil {
				return fmt.Errorf("locate the moai binary: %w", err)
			}
			child := exec.CommandContext(ctx, self, c.LaunchArgs()...)
			child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := factoryRelaunchExecFn(child); err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					return &exitCodeError{code: exitErr.ExitCode()}
				}
				return fmt.Errorf("relaunch %s: %w", c.LaunchLine(), err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&provider, factory.RelaunchFlagProvider, factory.RelaunchProviderCC, "Launcher to re-enter through: cc, glm, or codex")
	cmd.Flags().StringVar(&lane, factory.RelaunchFlagLane, "", "Lane label the stale session carried (lane-<n>); recorded for the notice only: the join takes the next free lane, since -l takes no argument (not available with codex)")
	cmd.Flags().StringVar(&run, factory.RelaunchFlagRun, "", "Run to join (--factory-run); omitted resolves the single active run (not available with codex)")
	cmd.Flags().StringVar(&fromRun, factory.RelaunchFlagFromRun, "", "Retire this run first when it is active and its owner is dead; otherwise leave it untouched")
	cmd.Flags().BoolVar(&dryRun, factory.RelaunchFlagDryRun, false, "Print the launch line and write nothing")
	return cmd
}
