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

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
	"github.com/modu-ai/moai-adk/internal/gitenv"
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

// afGateWaitDelay is how long to wait for the gate's output pipes after its context ended: a descendant that keeps
// them open must not hold the command past the timeout.
const afGateWaitDelay = 5 * time.Second

// afGateTimeout is the default time a gate may run.
const afGateTimeout = 20 * time.Minute

// afGate runs the gate command in the candidate worktree, under a timeout, in its own process group (killed whole
// when the timeout ends), with the repository-scoping variables removed from its environment and its output bounded
// by the merge queue's writer. A non-zero exit fails the card with the tail of what the command printed, with
// secrets in it redacted: the tail goes to the ledger and to the operator.
func afGate(gate string, timeout time.Duration) mergeq.Verify {
	return func(ctx context.Context, dir string) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sh", "-c", gate)
		cmd.Dir = dir
		cmd.Env = gitenv.Env()
		cmd.WaitDelay = afGateWaitDelay
		afGateProcessGroup(cmd)
		var out bytes.Buffer
		w := mergeq.NewLimitedWriter(&out)
		cmd.Stdout, cmd.Stderr = w, w
		if err := cmd.Run(); err != nil {
			tail := strings.TrimSpace(out.String())
			if len(tail) > afGateTail {
				tail = tail[len(tail)-afGateTail:]
			}
			tail, _ = core.Redact(tail)
			if ctx.Err() != nil {
				err = fmt.Errorf("%w (the gate ran past %s)", err, timeout)
			}
			return fmt.Errorf("gate %q: %w: %s", afRedactedGate(gate), err, tail)
		}
		return nil
	}
}

// afRedactedGate is the gate text as it may be recorded: a secret written into the command line is not kept.
func afRedactedGate(gate string) string {
	g, _ := core.Redact(gate)
	return g
}

// afCardMayLand says whether the card is where a landing needs it: MERGING at the epoch the queue row was made
// at. A holder who moved the card back (to REVIEWING, say) withdrew the work, and the row it left behind is not
// permission to land it. A row already landed answers landed on a retry and needs no position.
func afCardMayLand(ctx context.Context, st *ledger.Store, item mergeq.Item) error {
	if item.State == mergeq.StateLanded {
		return nil
	}
	cards, err := st.Cards(ctx)
	if err != nil {
		return err
	}
	for _, c := range cards {
		if c.ID != item.CardID {
			continue
		}
		if c.Position.Stage != core.StageMerging || c.Position.RemoteIntegrated || c.Epoch != item.CardEpoch {
			return fmt.Errorf("af land: %s is not in MERGING at the epoch its queue row was made at (stage %s, epoch %d; row epoch %d): "+
				"the holder moved it out of the merge queue, so it is not landed", item.CardID, c.Position.Stage, c.Epoch, item.CardEpoch)
		}
		return nil
	}
	return fmt.Errorf("af land: card %s does not exist", item.CardID)
}

// afReapDeadLocks expires the locks held by a process that is gone, so a landing lock whose holder died does not
// keep the queue for its whole ttl. It goes through Reap, but the judge it passes never calls a card's lease holder
// Dead (those pairs are answered Unknown, which the ledger never reclaims): only locks are taken back here; taking
// cards back is `moai af reap`.
func afReapDeadLocks(ctx context.Context, st *ledger.Store) error {
	leases, err := st.Leases(ctx)
	if err != nil {
		return err
	}
	type holder struct {
		pid   int
		start string
	}
	leased := map[holder]bool{}
	for _, l := range leases {
		leased[holder{l.PID, l.PIDStart}] = true
	}
	_, err = st.Reap(ctx, func(pid int, start string) ledger.Liveness {
		if leased[holder{pid, start}] {
			return ledger.LivenessUnknown
		}
		return afAlive(pid, start)
	})
	return err
}

func newAFLandCmd() *cobra.Command {
	var f afMergeFlags
	var gate string
	var gateTimeout time.Duration
	cmd := &cobra.Command{
		Use:   "land <card>",
		Short: "Land an enqueued card: verify the merge outside the lock, then move the trunk under it",
		Long: `Land an enqueued card.

The card's queued candidate commit is merged into the trunk in a throwaway worktree, the gate command runs there
(outside the landing lock), and only then is the landing lock taken and the trunk ref moved with a compare-and-swap.
A gate that fails fails the card with the tail of its output; a trunk that moved meanwhile is rebuilt and verified
again. The result is one JSON line (outcome: landed, failed or waiting).

The card must be in MERGING at the epoch its queue row was made at; a card its holder moved out of the merge queue
is refused. The gate runs with sh -c (native Windows has no sh: run this verb from WSL or Git Bash there), in its
own process group that is killed when --gate-timeout ends, with GIT_DIR and the other repository-scoping
variables removed from its environment. The gate text is recorded in the landed note, so do not put a secret in it.
A landing lock whose holder process is gone is taken back first.`,
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
				if err := afCardMayLand(ctx, st, item); err != nil {
					return err
				}
				if err := afReapDeadLocks(ctx, st); err != nil {
					return fmt.Errorf("af land: take back a dead holder's lock: %w", err)
				}
				m := afMerger(f, st, scratch, afGate(gate, gateTimeout))
				m.Gate = afRedactedGate(gate)
				res, err := m.Land(ctx, args[0], item.CandidateSHA)
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
	cmd.Flags().DurationVar(&gateTimeout, "gate-timeout", afGateTimeout, "how long the gate may run before its process group is killed")
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
				if err := afReapDeadLocks(ctx, st); err != nil {
					return fmt.Errorf("af reconcile: take back a dead holder's lock: %w", err)
				}
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
