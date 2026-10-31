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
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/flow"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/session"
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
	af.AddCommand(newAFClaimCmd(), newAFMoveCmd(), newAFEvidenceCmd(), newAFEnqueueCmd(), newAFReleaseCmd(), newAFReapCmd(), newAFLandCmd(), newAFReconcileCmd())
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

// afCheckCommit refuses a commit name that is not the complete lowercase hex SHA (40, or 64 for SHA-256) of a commit
// that exists in the repository: an abbreviation, HEAD or a branch name can name another commit tomorrow, so
// evidence or a queue entry bound to one is not bound to anything, and a value that starts with '-' reads as an
// option to git. The repository is --repo, else the worktree of the card's lease, else the project directory.
func afCheckCommit(ctx context.Context, st *ledger.Store, card, repoFlag, sha, what string) error {
	if !mergeq.IsCompleteSHA(sha) {
		return fmt.Errorf("af: %s %q is not a complete lowercase hex commit SHA (40 or 64 characters)", what, sha)
	}
	repo := repoFlag
	if repo == "" {
		if leases, err := st.Leases(ctx); err == nil {
			for _, l := range leases {
				if l.CardID == card && l.Worktree != "" {
					repo = l.Worktree
					break
				}
			}
		}
	}
	if repo == "" {
		repo = resolveProjectDir()
	}
	if got, err := (mergeq.GitCLI{Repo: repo}).RevParse(ctx, sha); err != nil || got != sha {
		return fmt.Errorf("af: %s %s names no commit in %s", what, sha, repo)
	}
	return nil
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

// afOwnerPID and afProbe are the seams of afHolder.
var (
	afOwnerPID = session.ResolveOwnerPID
	afProbe    = homestate.ProbeProcessIdentity
)

// afHolder is the process the lease records as the holder of the card. The command exits at once and is usually
// started through a shell wrapper that exits with it, so neither this process nor its parent can be the holder:
// a lease naming either reads as a dead holder at the next reap and the card is taken from a live session. The
// holder is the long-lived session that owns this process (the same resolution the session registry uses). When
// it cannot be resolved the holder is recorded as unknown (pid 0): the ledger never reclaims an unknown holder.
func afHolder() (pid int, start string) {
	owner, ok := afOwnerPID()
	if !ok || owner < 1 {
		return 0, ""
	}
	start, _ = afProbe(owner)
	if start == "" {
		// A pid without its fingerprint cannot be judged (a reused pid would pass for the session), so the
		// holder is recorded as unknown rather than leaving the verdict to the judge.
		return 0, ""
	}
	return owner, start
}

func newAFClaimCmd() *cobra.Command {
	var f afFlags
	var sessionID, worktree string
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
			pid, start := afHolder()
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				c, err := st.Claim(ctx, ledger.ClaimRequest{SessionID: sessionID, Worktree: worktree, PID: pid, PIDStart: start, Alive: afAlive})
				if err != nil {
					return err
				}
				if path, perr := afLedgerPath(f.db); perr == nil {
					// the session's hooks renew the lease from this binding; a failure only costs the renewal
					if berr := hook.BindLedgerHeartbeat(resolveProjectDir(), sessionID, hook.LedgerBinding{DB: path, CardID: c.CardID, Epoch: c.Epoch}); berr != nil {
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
	cmd.Flags().StringVar(&sessionID, "session", "", "the claiming session id")
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
	var repo string
	cmd := &cobra.Command{
		Use:   "evidence <card>",
		Short: "Record evidence (red, green, gate, review, merge) bound to a commit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				if err := afCheckCommit(ctx, st, args[0], repo, ev.SubjectSHA, "--subject-sha"); err != nil {
					return err
				}
				if err := st.AddEvidence(ctx, args[0], f.epoch, ev); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "kind": ev.Kind, "subject_sha": ev.SubjectSHA, "result": ev.Result})
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&ev.Kind, "kind", "", "red, green, gate, review or merge")
	cmd.Flags().StringVar(&ev.SubjectSHA, "subject-sha", "", "the complete lowercase hex SHA (40 or 64 characters) of the commit that was measured; it must exist in the repository")
	cmd.Flags().StringVar(&repo, "repo", "", "the git repository the commit is looked up in (default: the card's lease worktree, else the project directory)")
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
	var candidate, repo string
	cmd := &cobra.Command{
		Use:   "enqueue <card>",
		Short: "Put the card's candidate commit on the merge queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				if err := afCheckCommit(ctx, st, args[0], repo, candidate, "--candidate"); err != nil {
					return err
				}
				if err := st.Enqueue(ctx, args[0], f.epoch, candidate); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "candidate": candidate, "state": ledger.QueueWaiting})
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&candidate, "candidate", "", "the complete lowercase hex SHA (40 or 64 characters) of the commit the gates passed for; it must exist in the repository")
	cmd.Flags().StringVar(&repo, "repo", "", "the git repository the commit is looked up in (default: the card's lease worktree, else the project directory)")
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

// afAlive judges a lease holder by the identity the claim recorded (pid plus start fingerprint). Only a
// positively dead process is Dead; a missing pid or fingerprint, or a probe that cannot decide, is Unknown,
// which the ledger never reclaims (I5). It is the flow read model's adapter over homestate's classification,
// the one rule `claim` presence also uses.
func afAlive(pid int, start string) ledger.Liveness { return flow.ProcessProbe(afProbe)(pid, start) }

// newAFReapCmd is `moai af reap`: it takes back the cards whose holder process is gone (the ledger's Reap, judged
// by afAlive). A live holder keeps its card; the third reclaim of a card parks it for the operator.
func newAFReapCmd() *cobra.Command {
	var f afFlags
	cmd := &cobra.Command{
		Use:   "reap",
		Short: "Take back the cards whose holder process is gone",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return afRun(f.db, func(ctx context.Context, st *ledger.Store) error {
				got, err := st.Reap(ctx, afAlive)
				if err != nil {
					return err
				}
				if got == nil {
					got = []ledger.Reclaimed{}
				}
				leases, err := st.Leases(ctx)
				if err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"reclaimed": got, "unknown_expired": afUnknownExpired(leases, time.Now())})
			})
		},
	}
	f.bind(cmd, false)
	return cmd
}

// afUnknownLease is a lease whose holder is unknown and whose time has run out: the ledger will not reclaim it by
// liveness, so a lane that really died while holding it stays there until the operator looks.
type afUnknownLease struct {
	CardID    string    `json:"card_id"`
	Epoch     int64     `json:"epoch"`
	ExpiresAt time.Time `json:"expires_at"`
}

// afUnknownExpired lists the expired leases that carry no usable holder identity (no pid or no fingerprint).
func afUnknownExpired(leases []ledger.Lease, now time.Time) []afUnknownLease {
	out := []afUnknownLease{}
	for _, l := range leases {
		if (l.PID < 1 || l.PIDStart == "") && !l.ExpiresAt.After(now) {
			out = append(out, afUnknownLease{CardID: l.CardID, Epoch: l.Epoch, ExpiresAt: l.ExpiresAt})
		}
	}
	return out
}
