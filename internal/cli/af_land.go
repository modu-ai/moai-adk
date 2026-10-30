package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// afLandTimeout bounds `af land` and `af reconcile`: a landing runs the gates, which take far longer than a
// ledger write.
const afLandTimeout = 30 * time.Minute

// afGateTail is how much of a failed gate's output goes into the card's fail reason.
const afGateTail = 2048

// The merge queue's operational entry points (AC-AFL-014): `af land` lands an enqueued card through the queue's
// compare-and-swap of the trunk ref, and `af reconcile` settles the rows a restart left in verifying or landing
// by git ancestry. Both are thin: mergeq.Merger owns every rule.

// afMergeFlags are the flags both verbs take.
type afMergeFlags struct {
	db, repo, trunk, session string
}

func (f *afMergeFlags) bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.db, "db", "", "ledger file (default: <project dir>/af/ledger.db)")
	cmd.Flags().StringVar(&f.repo, "repo", "", "the git repository the trunk lives in (default: the project directory)")
	cmd.Flags().StringVar(&f.trunk, "trunk", mergeq.DefaultTrunk, "the trunk ref, a full refs/heads/ name checked out in no worktree")
	cmd.Flags().StringVar(&f.session, "session", "moai-af", "who holds the landing lock (recorded)")
}

// afMerger builds the merge queue over the opened ledger. verify may be nil for a verb that runs no gates.
func afMerger(f afMergeFlags, st *ledger.Store, scratch string, verify mergeq.Verify) *mergeq.Merger {
	repo := f.repo
	if repo == "" {
		repo = resolveProjectDir()
	}
	pid := os.Getpid() // the lock is held while this command runs, so this process is the holder
	start, _ := homestate.ProbeProcessIdentity(pid)
	return &mergeq.Merger{
		Git:    mergeq.GitCLI{Repo: repo, ScratchD: scratch, Who: mergeq.Identity{Name: "moai af", Email: "af@moai.invalid"}},
		Queue:  mergeq.NewLedgerQueue(st),
		Trunk:  f.trunk,
		Verify: verify, Holder: mergeq.Holder{Session: f.session, PID: pid, PIDStart: start},
		LockTTL: mergeq.DefaultLockTTL, MaxRetries: mergeq.DefaultMaxRetries,
		Message: func(card string) string { return "merge(" + card + "): land" },
	}
}

// afGate runs the gate command in the candidate worktree. A non-zero exit fails the card with the tail of what the
// command printed.
func afGate(gate string) mergeq.Verify {
	return func(ctx context.Context, dir string) error {
		cmd := exec.CommandContext(ctx, "sh", "-c", gate)
		cmd.Dir = dir
		var out bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &out
		if err := cmd.Run(); err != nil {
			tail := strings.TrimSpace(out.String())
			if len(tail) > afGateTail {
				tail = tail[len(tail)-afGateTail:]
			}
			return fmt.Errorf("gate %q: %w: %s", gate, err, tail)
		}
		return nil
	}
}

func newAFLandCmd() *cobra.Command {
	var f afMergeFlags
	var gate string
	cmd := &cobra.Command{
		Use:   "land <card>",
		Short: "Land an enqueued card: verify the merge outside the lock, then move the trunk under it",
		Long: `Land an enqueued card.

The card's queued candidate commit is merged into the trunk in a throwaway worktree, the gate command runs there
(outside the landing lock), and only then is the landing lock taken and the trunk ref moved with a compare-and-swap.
A gate that fails fails the card with the tail of its output; a trunk that moved meanwhile is rebuilt and verified
again. The result is one JSON line (outcome: landed, failed or waiting).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(gate) == "" {
				return fmt.Errorf("af land: --gate is required: a landing without gates would merge unverified work")
			}
			return afMergeRun(cmd, f, func(ctx context.Context, st *ledger.Store, scratch string) error {
				q := mergeq.NewLedgerQueue(st)
				item, err := q.Item(ctx, args[0], time.Now())
				if err != nil {
					return fmt.Errorf("af land: %w", err)
				}
				if item.State == mergeq.StateFailed {
					return fmt.Errorf("af land: %s failed in the queue (queue it again with enqueue)", args[0])
				}
				res, err := afMerger(f, st, scratch, afGate(gate)).Land(ctx, args[0], item.CandidateSHA)
				if err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{
					"card_id": args[0], "outcome": res.Outcome, "reason": res.Reason,
					"trunk_before": res.TrunkBefore, "trunk_after": res.TrunkAfter, "verified": res.Verified,
				})
			})
		},
	}
	f.bind(cmd)
	cmd.Flags().StringVar(&gate, "gate", "", "the command that verifies the merge result, run with sh -c in the candidate worktree")
	return cmd
}

func newAFReconcileCmd() *cobra.Command {
	var f afMergeFlags
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Settle the queue rows a restart left in verifying or landing, by git ancestry",
		Long: `Settle the queue rows a restart left behind.

A row whose candidate is already in the trunk is settled as landed; one whose candidate is not goes back to
waiting (an older landing row is settled as failed), and only while this command holds the landing lock. A verifying
row whose card still holds a live lease is left alone. Every row is judged on its own: one failure does not stop the
others. The result is one JSON line listing the actions taken; a failed row is reported as an error after it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return afMergeRun(cmd, f, func(ctx context.Context, st *ledger.Store, scratch string) error {
				items, err := mergeq.NewLedgerQueue(st).Items(ctx, time.Now())
				if err != nil {
					return err
				}
				acts, rerr := afMerger(f, st, scratch, nil).Reconcile(ctx, items)
				out := make([]map[string]string, 0, len(acts))
				for _, a := range acts {
					out = append(out, map[string]string{"card_id": a.CardID, "from": string(a.From), "to": string(a.To), "why": a.Why})
				}
				if perr := afPrint(cmd, map[string]any{"actions": out}); perr != nil {
					return perr
				}
				return rerr
			})
		},
	}
	f.bind(cmd)
	return cmd
}

// afMergeRun opens the ledger and a scratch directory for the candidate worktrees, runs fn under the landing
// timeout and cleans up.
func afMergeRun(cmd *cobra.Command, f afMergeFlags, fn func(ctx context.Context, st *ledger.Store, scratch string) error) error {
	path, err := afLedgerPath(f.db)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), afLandTimeout)
	defer cancel()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		return fmt.Errorf("af: open the ledger %s: %w", path, err)
	}
	defer func() { _ = st.Close() }()
	scratch, err := os.MkdirTemp("", "moai-af-land-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	return fn(ctx, st, scratch)
}
