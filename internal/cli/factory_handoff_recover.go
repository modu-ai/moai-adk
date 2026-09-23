package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// abandonLaneProbe is the source-owner liveness probe `abandon-lane` passes to
// the broker's terminal transaction. It is the only seam: tests swap this
// variable, never a store field or a broker row.
var abandonLaneProbe = homestate.ProbeProcessIdentity

func newFactoryCommand() *cobra.Command {
	factory := &cobra.Command{Use: "factory", Short: "Factory runtime maintenance"}
	handoff := &cobra.Command{Use: "handoff", Short: "Factory handoff maintenance"}
	var id int64
	var token, decision string
	recover := &cobra.Command{Use: "recover-resume", Short: "Recover an indeterminate legacy resume claim", RunE: func(cmd *cobra.Command, _ []string) error {
		if id <= 0 || token == "" || (decision != "requeue" && decision != "fail") {
			return fmt.Errorf("--id, --expected-token, and --decision requeue|fail are required")
		}
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		db, err := homestate.OpenFactory(root)
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
		return db.RecoverLegacyResume(context.Background(), id, token, decision, homestate.ProbeProcessIdentity, func() error {
			census, err := homestate.ReadRuntimeCensus(root)
			if err != nil {
				return err
			}
			if census.Total() != 0 {
				return fmt.Errorf("active runtimes=%d", census.Total())
			}
			return nil
		})
	}}
	recover.Flags().Int64Var(&id, "id", 0, "Resume handoff row id")
	recover.Flags().StringVar(&token, "expected-token", "", "Current claim token")
	recover.Flags().StringVar(&decision, "decision", "", "Recovery decision: requeue or fail")
	handoff.AddCommand(recover)
	handoff.AddCommand(newAbandonLaneCommand())
	factory.AddCommand(handoff)
	return factory
}

func init() { rootCmd.AddCommand(newFactoryCommand()) }

// newAbandonLaneCommand is the operator's manual recovery for a lane stuck
// behind a non-final handoff (REQ-FLH-011): it terminates the handoff as
// ABANDONED/OPERATOR_ABANDONED only when the recorded source owner is not
// current, and never touches the target worktree, its branch, or its commits.
// There is no time-based automatic termination.
func newAbandonLaneCommand() *cobra.Command {
	var slot, run string
	cmd := &cobra.Command{
		Use:   "abandon-lane",
		Short: "Terminate a lane's stuck non-final worktree handoff (source owner must not be current)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if slot == "" {
				return fmt.Errorf("--slot is required")
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			primary := homestate.CanonicalProjectRoot(cwd)
			runID, err := factorymsg.ResolveActiveRun(ctx, primary, run)
			if err != nil {
				return err
			}
			store, err := factorymsg.Open(primary, runID)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			h, err := store.AbandonLane(ctx, slot, abandonLaneProbe)
			if err != nil {
				if reason, ok := factorymsg.HandoffNackReason(err); ok {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: lane %s: %s; nothing was written\n", slot, reason)
				}
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "moai: lane %s handoff %s is %s (%s): card=%s spec=%s; worktree %s and branch %s are preserved\n",
				slot, h.ID, h.State, h.Reason, h.CardID, h.SpecID, h.TargetPath, h.TargetBranch)
			return nil
		},
	}
	cmd.Flags().StringVar(&slot, "slot", "", "Lane slot whose handoff to abandon")
	cmd.Flags().StringVar(&run, "run", "", "Factory run id (default: the single active run)")
	return cmd
}
