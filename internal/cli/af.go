package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// afOpTimeout bounds one verb: the ledger write is short and a verb never waits on a person.
const afOpTimeout = 30 * time.Second

// newAFCmd is the `moai af` command: the verbs of the card flow over the local ledger (AC-AFL-017). One verb per
// ledger method the flow needs, and nothing else: claim, move (stage move), evidence (record), enqueue, release.
// Each verb opens the ledger, makes its one call, prints one JSON object and exits; the ledger owns every rule
// (fencing epochs, the position table, the evidence binding), so a verb adds none. Landing is the merge queue's
// and has no verb here.
func newAFCmd() *cobra.Command {
	af := &cobra.Command{
		Use:          "af",
		Short:        "Agentic factory card flow over the local ledger",
		SilenceUsage: true,
	}
	af.AddCommand(newAFClaimCmd(), newAFMoveCmd(), newAFEvidenceCmd(), newAFEnqueueCmd(), newAFReleaseCmd())
	return af
}

// afLedgerPath is the ledger file of this project: <project dir>/af/ledger.db, the location the ledger contract
// names. --db overrides it (tests, or a ledger kept elsewhere).
func afLedgerPath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	dir, err := homestate.ProjectDir(resolveProjectDir())
	if err != nil {
		return "", fmt.Errorf("af: cannot resolve the project directory: %w", err)
	}
	return filepath.Join(dir, "af", "ledger.db"), nil
}

// afRun opens the ledger, runs fn with a bounded context and closes the ledger.
func afRun(db string, fn func(ctx context.Context, st *ledger.Store) error) error {
	path, err := afLedgerPath(db)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), afOpTimeout)
	defer cancel()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		return fmt.Errorf("af: open the ledger %s: %w", path, err)
	}
	defer func() { _ = st.Close() }()
	return fn(ctx, st)
}

func afPrint(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	return enc.Encode(v)
}

// afFlags are the flags every verb carries.
type afFlags struct {
	db    string
	epoch int64
}

func (f *afFlags) bind(cmd *cobra.Command, needEpoch bool) {
	cmd.Flags().StringVar(&f.db, "db", "", "ledger file (default: <project dir>/af/ledger.db)")
	if needEpoch {
		cmd.Flags().Int64Var(&f.epoch, "epoch", 0, "the fencing epoch the claim returned")
		_ = cmd.MarkFlagRequired("epoch")
	}
}

func newAFClaimCmd() *cobra.Command {
	var f afFlags
	var session, worktree string
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Claim the next card and print its id and fencing epoch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if worktree == "" {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				worktree = wd
			}
			// The lease records the process that holds the card. This command exits at once, so the holder is the
			// session that ran it: the parent process.
			pid := os.Getppid()
			start, _ := homestate.ProbeProcessIdentity(pid)
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				c, err := st.Claim(ctx, ledger.ClaimRequest{SessionID: session, Worktree: worktree, PID: pid, PIDStart: start})
				if err != nil {
					return err
				}
				if path, perr := afLedgerPath(f.db); perr == nil {
					// the session's hooks renew the lease from this binding; a failure only costs the renewal
					if berr := hook.BindLedgerHeartbeat(resolveProjectDir(), session, hook.LedgerBinding{DB: path, CardID: c.CardID, Epoch: c.Epoch}); berr != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "af: the lease will not be renewed by hooks: %v\n", berr)
					}
				}
				return afPrint(cmd, map[string]any{
					"card_id": c.CardID, "epoch": c.Epoch,
					"stage": c.Position.Stage, "sub": c.Position.Sub, "expires_at": c.ExpiresAt,
				})
			})
		},
	}
	f.bind(cmd, false)
	cmd.Flags().StringVar(&session, "session", "", "the claiming session id")
	cmd.Flags().StringVar(&worktree, "worktree", "", "the worktree the card is worked in (default: the working directory)")
	_ = cmd.MarkFlagRequired("session")
	return cmd
}

func newAFMoveCmd() *cobra.Command {
	var f afFlags
	var stage, sub, actor, reason string
	var remote bool
	cmd := &cobra.Command{
		Use:   "move <card>",
		Short: "Move a claimed card to a stage",
		Long: `Move a claimed card to a stage.

Limits: the verb carries no guard or evidence arguments. A move that needs a guard key (an operator hold, a
fix-limit or reclaim-cap park, a resume guard) or evidence handed to the position table is refused by the
ledger, and this verb cannot supply either; such moves are made through the ledger API. Recording evidence is
the separate "evidence" verb.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			to := core.Position{Stage: core.Stage(strings.ToUpper(stage)), Sub: core.SubStage(strings.ToUpper(sub)), RemoteIntegrated: remote}
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				if err := st.Move(ctx, args[0], f.epoch, to, actor, reason); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "stage": to.Stage, "sub": to.Sub})
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&stage, "stage", "", "target stage: READY, BUILDING, REVIEWING, MERGING, DONE or CANCELLED")
	cmd.Flags().StringVar(&sub, "sub", "", "target sub-stage within BUILDING: SPEC, IMPLEMENT or FIX")
	cmd.Flags().BoolVar(&remote, "remote-integrated", false, "MERGING with the merge seen")
	cmd.Flags().StringVar(&actor, "actor", "lane", "who makes the move (recorded)")
	cmd.Flags().StringVar(&reason, "reason", "", "why (recorded)")
	_ = cmd.MarkFlagRequired("stage")
	return cmd
}

func newAFEvidenceCmd() *cobra.Command {
	var f afFlags
	var ev ledger.Evidence
	cmd := &cobra.Command{
		Use:   "evidence <card>",
		Short: "Record evidence (red, green, gate, review, merge) bound to a commit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				if err := st.AddEvidence(ctx, args[0], f.epoch, ev); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "kind": ev.Kind, "subject_sha": ev.SubjectSHA, "result": ev.Result})
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&ev.Kind, "kind", "", "red, green, gate, review or merge")
	cmd.Flags().StringVar(&ev.SubjectSHA, "subject-sha", "", "the commit SHA that was measured")
	cmd.Flags().StringVar(&ev.TreeSHA, "tree-sha", "", "the tree SHA (informational)")
	cmd.Flags().StringVar(&ev.BaseSHA, "base-sha", "", "the base commit SHA")
	cmd.Flags().StringVar(&ev.Command, "command", "", "the command that was run")
	cmd.Flags().StringVar(&ev.Result, "result", "", "pass or fail")
	cmd.Flags().StringVar(&ev.Path, "path", "", "a file holding the output")
	_ = cmd.MarkFlagRequired("kind")
	_ = cmd.MarkFlagRequired("subject-sha")
	_ = cmd.MarkFlagRequired("result")
	return cmd
}

func newAFEnqueueCmd() *cobra.Command {
	var f afFlags
	var candidate string
	cmd := &cobra.Command{
		Use:   "enqueue <card>",
		Short: "Put the card's candidate commit on the merge queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				if err := st.Enqueue(ctx, args[0], f.epoch, candidate); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "candidate": candidate, "state": ledger.QueueWaiting})
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&candidate, "candidate", "", "the commit SHA the gates passed for")
	_ = cmd.MarkFlagRequired("candidate")
	return cmd
}

func newAFReleaseCmd() *cobra.Command {
	var f afFlags
	cmd := &cobra.Command{
		Use:   "release <card>",
		Short: "Let go of a claimed card",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				if err := st.Release(ctx, args[0], f.epoch); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "released": true})
			})
		},
	}
	f.bind(cmd, true)
	return cmd
}
