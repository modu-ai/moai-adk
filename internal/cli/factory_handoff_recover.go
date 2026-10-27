package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// abandonLaneProbe is the source-owner liveness probe `abandon-lane` passes to
// the broker's terminal transaction. It is the only seam: tests swap this
// variable, never a store field or a broker row.
var abandonLaneProbe = homestate.ProbeProcessIdentity

// verifyWorkingDir returns the working-directory error the project-dir
// resolution swallowed, so a verb never opens a factory at a fabricated
// root: dir may be empty (the resolution's unresolved branch) or name a
// path that no longer exists — os.Getwd can still succeed with a deleted
// working directory, and OpenFactory would then recreate the tree under
// it instead of failing (Codex PR #1833 review P2-1, card t1862).
func verifyWorkingDir(dir string) error {
	if dir == "" {
		if _, err := os.Getwd(); err != nil {
			return err
		}
		return fmt.Errorf("project directory unresolved")
	}
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("working directory %s is unavailable: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("working directory %s is not a directory", dir)
	}
	return nil
}

// checkedFactoryRoot validates the launch directory before applying the shared
// queue-root rule, so maintenance verbs cannot open a fabricated factory root.
// @MX:ANCHOR: [AUTO] Checked root resolution for factory maintenance verbs.
// @MX:REASON: Recovery, run listing, and relaunch must share validation and mapping.
func checkedFactoryRoot() (string, error) {
	dir, _ := resolveProjectDirWithSource()
	if err := verifyWorkingDir(dir); err != nil {
		return "", err
	}
	return factoryRootOf(dir), nil
}

func newFactoryCommand() *cobra.Command {
	factory := &cobra.Command{Use: "factory", Short: "Factory runtime maintenance", Long: "Factory runtime maintenance.\n\nRun `moai paths` for the confirmed factory and broker database paths; never guess\na database path. Opening a missing file with sqlite3 creates an empty one."}
	handoff := &cobra.Command{Use: "handoff", Short: "Factory handoff maintenance"}
	var id int64
	var token, decision string
	recover := &cobra.Command{Use: "recover-resume", Short: "Recover an indeterminate legacy resume claim", RunE: func(cmd *cobra.Command, _ []string) error {
		if id <= 0 || token == "" || (decision != "requeue" && decision != "fail") {
			return fmt.Errorf("--id, --expected-token, and --decision requeue|fail are required")
		}
		root, err := checkedFactoryRoot()
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
	handoff.AddCommand(newFactoryAdoptCommand())
	factory.AddCommand(handoff)
	factory.AddCommand(newFactoryRunsCommand(), newFactoryRelaunchCommand())
	factory.AddCommand(newFactoryAssignCommand(), newFactoryBundleCommand(), newFactoryStatusCommand(), newFactoryDecideCommand(),
		newFactoryNextCommand(), newFactoryStageCommand(), newFactoryCompleteCommand(), newFactoryRenewCommand())
	factory.AddCommand(newFactorySetSpecCommand(), newFactoryReclassifyCommand())
	factory.AddCommand(newFactoryRebindEvidenceCommand()) // card t1930 defect ④
	factory.AddCommand(newFactoryApproveCommand())
	factory.AddCommand(newFactoryMessagingCommand(), newFactoryFallbackCommand())
	factory.AddCommand(newFactoryPickupCommand())
	factory.AddCommand(newFactoryMergeCommand())
	return factory
}

// newFactoryRunsCommand is the operator maintenance surface (REQ-008). It
// exists because a run whose owner is indeterminate is never auto-retired by
// design, so without a deliberate operator surface such a row would be
// permanently unremovable.
//
// It shares the reconciler's predicate rather than carrying a second, laxer
// copy of it: retirement happens only on a positive `dead` classification, on
// this path exactly as on every other (REQ-005). There is no interactive
// prompt — the subagent boundary forbids it, and the command is operator-invoked.
//
// A run that still holds in-progress cards is refused, with the cards and the
// exits listed (SPEC-FACTORY-RETIRE-GUARD-001); the refusal is a temporary
// safeguard until retirement can transfer the cards. Retiring an already-retired
// run succeeds without a write and lists the cards it still strands.
//
// With --rebind-to the retirement becomes a consolidation (card t1963,
// SPEC-FACTORY-COMPLETION-RECOVERY-001 REQ-FCR-021): the dead run's
// in-progress cards move under the named live run and the dead run retires, in
// one transaction. It is the leader and operator path only — a lane session is
// refused — and without the flag nothing about the retirement changes.
func newFactoryRunsCommand() *cobra.Command {
	var retire, rebindTo string
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Report factory runs with their owner liveness, and retire a dead-owner run",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			if rebindTo != "" {
				if retire == "" {
					return fmt.Errorf("factory runs: --rebind-to needs --retire <dead-run>: it names where the retired run's cards go, and nothing was changed")
				}
				if factoryLaneRefusal() {
					return fmt.Errorf("factory runs --rebind-to: refused — %s: consolidating runs is the leader and operator path's act, not a lane session's; nothing was changed", factoryLaneBoundarySentinel)
				}
			}
			root, err := checkedFactoryRoot()
			if err != nil {
				return err
			}
			db, err := homestate.OpenFactory(root)
			if err != nil {
				return err
			}
			defer closeFactoryInto(&err, db, "factory state")
			opts := factorymsg.ReconcileOptionsFor(root)
			if rebindTo != "" {
				return runRebindRetire(cmd, db, opts, retire, rebindTo)
			}
			if retire != "" {
				classification, retireErr := db.RetireRunIfDead(cmd.Context(), retire, opts)
				var cardsErr *homestate.RunInProgressCardsError
				if errors.As(retireErr, &cardsErr) {
					// Refuse-only: list the cards and the exits, retire nothing.
					writeRetireRefusal(cmd.OutOrStdout(), retire, classification, cardsErr.Cards)
					return retireErr
				}
				if retireErr != nil {
					return retireErr
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "retired %s (owner %s)\n", retire, classification)
				// Read after the retirement call returned, so no transaction is open.
				stranded, err := db.InProgressCards(cmd.Context(), retire)
				if err != nil {
					return err
				}
				writeStrandedCards(cmd.OutOrStdout(), retire, stranded)
				return nil
			}
			owners, err := db.ClassifyRuns(cmd.Context(), opts)
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "RUN ID\tSTATUS\tOWNER\tPID\tCREATED\tUPDATED")
			for _, o := range owners {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n", o.RunID, o.Status, o.Classification, o.LeadPID, o.CreatedAt, o.UpdatedAt)
			}
			return w.Flush()
		},
	}
	cmd.Flags().StringVar(&retire, "retire", "", "Retire the named run — refused unless its owner classifies dead and it holds no in-progress cards")
	cmd.Flags().StringVar(&rebindTo, "rebind-to", "", "With --retire: move the dead run's in-progress cards under this live run, then retire it (leader and operator only)")
	return cmd
}

// runRebindRetire is `runs --retire <source> --rebind-to <target>`: it
// classifies both runs, hands them to the consolidation operation and prints
// the cards that moved and the ones abandoned with a pointer. Every refusal is
// the operation's own — an owner that is not dead, a target that is not live, a
// live worker, a card another live run holds — and changes nothing.
func runRebindRetire(cmd *cobra.Command, db *homestate.FactoryDB, opts homestate.ReconcileOptions, source, target string) error {
	owners, err := db.ClassifyRuns(cmd.Context(), opts)
	if err != nil {
		return err
	}
	var src, tgt *homestate.RunOwner
	for i := range owners {
		switch owners[i].RunID {
		case source:
			src = &owners[i]
		case target:
			tgt = &owners[i]
		}
	}
	if src == nil {
		return fmt.Errorf("factory run %s not found", source)
	}
	if tgt == nil {
		return fmt.Errorf("factory run %s not found (the --rebind-to target)", target)
	}
	res, err := db.RebindRun(cmd.Context(), homestate.RebindRequest{Source: *src, Target: *tgt}, opts)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "rebound %s into %s (owner %s)\n", source, target, src.Classification)
	_, _ = fmt.Fprintf(out, "moved cards: %s\n", joinOrNone(res.Moved))
	_, _ = fmt.Fprintf(out, "abandoned cards: %s\n", joinOrNone(res.Abandoned))
	return nil
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
